package updater

import (
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
	// ExecPath 待替换/待实测的可执行文件路径, 空串取 os.Executable() 自身 (测试与特殊布局可覆盖)
	ExecPath string
}

// UpdateResult 更新执行结果
type UpdateResult struct {
	From     string // 原版本展示串
	To       string // 安装后的版本 tag
	Method   Method // 实际走的更新通道
	UpToDate bool   // 未做任何变更, 当前已是最新
	// GoMissingFallback go install 构建委托时因 go 不在 PATH 而回退下载替换路径时为真
	GoMissingFallback bool
}

// Update 执行端到端自更新, 按安装方式分发:
// - go install 构建: 委托 go install <module>@<ver>, go 不在 PATH 时回退下载替换路径
// - release/本地构建: 查 Releases API 下载对应平台产物做替换
// 任一步失败旧版原位不动并清理临时产物; 已是最新时不下载直接返回 UpToDate。
func Update(opts UpdateOptions) (UpdateResult, error) {
	execPath, err := selfExecPath(opts.ExecPath)
	if err != nil {
		return UpdateResult{}, err
	}

	// 上次更新遗留的 .old 备份清理 (Windows 上运行中的旧映像无法删除, 顺延到本次启动)。
	// 放在分发层使 go install 委托路径同样清理遗留
	removeStaleOldBinary(execPath)

	if opts.Current.Method == MethodGoInstall {
		return updateGoInstall(opts, execPath)
	}
	return updateDownload(opts, execPath)
}

// selfExecPath 解析待操作的可执行文件路径: 显式指定优先, 否则取 os.Executable() 自身
func selfExecPath(explicit string) (string, error) {
	execPath := explicit
	if execPath == "" {
		p, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("定位自身可执行文件失败: %w", err)
		}
		execPath = p
	}
	if p, err := filepath.Abs(execPath); err == nil {
		execPath = p
	}
	return execPath, nil
}

// updateDownload release/本地构建用户的下载替换路径:
// 查 Releases API 选平台产物 → 下载 → sha256 校验 → 解出二进制 → 冒烟测试 → rename 原子交换。
// sha256 校验是安装的硬性前提: Release 缺 checksums.txt 时直接拒绝, 旧版原位不动
func updateDownload(opts UpdateOptions, execPath string) (UpdateResult, error) {
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
	if !hasChecksums {
		return UpdateResult{}, fmt.Errorf("版本 %s 未提供 checksums.txt, 无法校验产物完整性, 已拒绝安装 (旧版保持不变)", rel.TagName)
	}

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
	checksumsPath := filepath.Join(tmpDir, checksumsAsset.Name)
	if err := downloadFile(checksumsAsset.BrowserDownloadURL, checksumsPath); err != nil {
		return UpdateResult{}, err
	}
	if err := verifyChecksum(archivePath, checksumsPath, archive.Name); err != nil {
		return UpdateResult{}, err
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
		From:   opts.Current.Version,
		To:     rel.TagName,
		Method: opts.Current.Method,
	}, nil
}
