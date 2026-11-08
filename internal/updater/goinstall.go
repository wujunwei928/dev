package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// updateGoInstall go install 构建用户的更新路径:
// 裸更新先经 go list 查询最新版做 already-up-to-date 判定 (已是最新则不动),
// 再委托 go install <module>@latest (显式 --version 时 @vX.Y.Z) 完成升级,
// 完成后实测自身 --version 验证版本真的变了 (不信子进程退出码);
// go 不在 PATH 时回退到下载替换路径。
func updateGoInstall(opts UpdateOptions, execPath string) (UpdateResult, error) {
	// 目标版本: 显式 --version 按请求安装 (兼做降级回滚); 裸更新先查最新版做 already-up-to-date 判定
	var want string
	if opts.Target != "" {
		want = "v" + normalizeVersion(opts.Target)
	} else {
		chk, err := Check(opts.Current, "")
		if err != nil {
			if isGoMissingErr(err) {
				return fallbackToDownload(opts, execPath)
			}
			return UpdateResult{}, err
		}
		if chk.UpToDate {
			return UpdateResult{
				From:     opts.Current.Version,
				To:       chk.Latest,
				Method:   MethodGoInstall,
				UpToDate: true,
			}, nil
		}
	}

	ref := module + "@latest"
	if want != "" {
		ref = module + "@" + want
	}
	if err := runGoInstall(ref); err != nil {
		if isGoMissingErr(err) {
			return fallbackToDownload(opts, execPath)
		}
		return UpdateResult{}, err
	}

	// 事后实测: 不信 go install 的退出码, 版本必须真的变化
	got, err := measuredVersion(execPath)
	if err != nil {
		return UpdateResult{}, err
	}
	if want == "" {
		// 裸更新: 版本必须与原来不同
		if normalizeVersion(got) == normalizeVersion(opts.Current.Version) {
			return UpdateResult{}, fmt.Errorf("go install 已执行, 但实测版本未变: 仍为 %s, 安装可能未生效", got)
		}
	} else if normalizeVersion(got) != normalizeVersion(want) {
		// 指定版本: 实测必须与请求一致
		return UpdateResult{}, fmt.Errorf("go install 已执行, 但实测版本为 %s, 期望 %s, 安装可能未生效", got, want)
	}
	return UpdateResult{
		From:   opts.Current.Version,
		To:     got,
		Method: MethodGoInstall,
	}, nil
}

// runGoInstall 执行 go install <ref>, 构建输出直通用户终端
func runGoInstall(ref string) error {
	// #nosec G204 -- go 命令与版本 ref 均由常量模块路径和归一化版本号构成, 无任意用户输入
	cmd := exec.Command("go", "install", ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if isGoMissingErr(err) {
			return fmt.Errorf("未找到 go 命令: %w", err)
		}
		return fmt.Errorf("执行 go install %s 失败: %w", ref, err)
	}
	return nil
}

// measuredVersion 实测可执行文件 --version, 从输出解析版本号
// (任一空白分隔 token 可解析为 semver 即取, cobra 默认输出形如 "dev version 0.2.1")
func measuredVersion(execPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), smokeTestTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, execPath, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("go install 已执行, 但实测 %s --version 出错: %v, 安装可能未生效", execPath, err)
	}
	for _, f := range strings.Fields(string(out)) {
		if _, ok := parseSemver(f); ok {
			return "v" + normalizeVersion(f), nil
		}
	}
	return "", fmt.Errorf("go install 已执行, 但无法从 %s --version 输出 %q 解析出版本号, 安装可能未生效",
		execPath, strings.TrimSpace(string(out)))
}

// fallbackToDownload go 不可用时回退下载替换路径, 结果标注回退事实
func fallbackToDownload(opts UpdateOptions, execPath string) (UpdateResult, error) {
	res, err := updateDownload(opts, execPath)
	if err != nil {
		return UpdateResult{}, err
	}
	res.GoMissingFallback = true
	return res, nil
}

// isGoMissingErr 报告 err 是否由 go 命令不存在 (不在 PATH) 引起
func isGoMissingErr(err error) bool {
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return errors.Is(execErr.Err, exec.ErrNotFound)
	}
	return errors.Is(err, exec.ErrNotFound)
}
