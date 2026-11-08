package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// UpdateOptions 更新执行参数
type UpdateOptions struct {
	// Current 当前版本与安装方式 (经 ResolveVersion/CurrentVersion 得出)
	Current VersionInfo
	// Target 目标版本 tag ("v0.2.0" / "0.2.0"), 空串表示最新版。
	// 显式指定时按请求安装 (即使当前更新, 兼做降级回滚)
	Target string
	// APIBase GitHub API base, 空串使用默认 https://api.github.com
	APIBase string
	// ExecPath 待替换的可执行文件路径, 空串取 os.Executable() 自身 (测试与特殊布局可覆盖)
	ExecPath string
}

// UpdateResult 更新执行结果
type UpdateResult struct {
	From             string // 原版本展示串
	To               string // 安装后的版本 tag
	Method           Method // 实际走的更新通道
	UpToDate         bool   // 未做任何变更, 当前已是最新
	ChecksumVerified bool   // 本次安装是否经过 checksums.txt sha256 校验 (历史 Release 未提供该文件时为 false)
}

// Update 执行 release/本地构建用户的端到端自更新:
// 查 Releases API 选平台产物 → 下载 → sha256 校验 → 解出二进制 → 冒烟测试 → rename 原子交换。
// 任一步失败旧版原位不动并清理临时产物; 已是最新时不下载直接返回 UpToDate。
// go install 构建不走本路径 (由 go install 委托负责)
func Update(opts UpdateOptions) (UpdateResult, error) {
	if opts.Current.Method == MethodGoInstall {
		return UpdateResult{}, errors.New("go install 构建请使用 go install 命令升级, 不走下载替换路径")
	}

	execPath := opts.ExecPath
	if execPath == "" {
		p, err := os.Executable()
		if err != nil {
			return UpdateResult{}, fmt.Errorf("定位自身可执行文件失败: %w", err)
		}
		execPath = p
	}
	if p, err := filepath.Abs(execPath); err == nil {
		execPath = p
	}

	// 上次更新遗留的 .old 备份清理 (Windows 上运行中的旧映像无法删除, 顺延到本次启动)
	removeStaleOldBinary(execPath)

	// 1. 查询目标 Release
	rel, err := fetchRelease(opts.APIBase, opts.Target)
	if err != nil {
		return UpdateResult{}, err
	}

	// 2. 已是最新判定 (仅裸更新; 显式 --version 按请求安装, 兼做降级)
	if opts.Target == "" {
		if cmp, ok := compareVersions(opts.Current.Version, rel.TagName); ok && cmp >= 0 {
			return UpdateResult{
				From:     opts.Current.Version,
				To:       rel.TagName,
				Method:   opts.Current.Method,
				UpToDate: true,
			}, nil
		}
	}

	// 3. 选择平台产物与校验文件
	archive, err := pickAsset(rel, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return UpdateResult{}, err
	}
	checksumsAsset, hasChecksums := findChecksumsAsset(rel)

	// 4. 下载到系统临时目录并校验 sha256
	tmpDir, err := os.MkdirTemp("", "dev-update-")
	if err != nil {
		return UpdateResult{}, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archive.Name)
	if err := downloadFile(archive.BrowserDownloadURL, archivePath); err != nil {
		return UpdateResult{}, err
	}
	checksumVerified := false
	if hasChecksums {
		checksumsPath := filepath.Join(tmpDir, checksumsAsset.Name)
		if err := downloadFile(checksumsAsset.BrowserDownloadURL, checksumsPath); err != nil {
			return UpdateResult{}, err
		}
		if err := verifyChecksum(archivePath, checksumsPath, archive.Name); err != nil {
			return UpdateResult{}, err
		}
		checksumVerified = true
	}

	// 5. 解出二进制, 暂存目录与目标同卷以保证 rename 交换
	stageDir, err := os.MkdirTemp(filepath.Dir(execPath), "dev-update-")
	if err != nil {
		return UpdateResult{}, fmt.Errorf("创建暂存目录失败: %w", err)
	}
	defer os.RemoveAll(stageDir)
	binPath := filepath.Join(stageDir, binaryName())
	if err := extractBinary(archivePath, binPath); err != nil {
		return UpdateResult{}, err
	}

	// 6. 冒烟测试: 新二进制 --version 必须等于目标版本
	if err := smokeTest(binPath, rel.TagName); err != nil {
		return UpdateResult{}, err
	}

	// 7. rename 原子交换
	if err := swapBinary(execPath, binPath); err != nil {
		return UpdateResult{}, err
	}

	return UpdateResult{
		From:             opts.Current.Version,
		To:               rel.TagName,
		Method:           opts.Current.Method,
		ChecksumVerified: checksumVerified,
	}, nil
}
