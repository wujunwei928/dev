package updater

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	// module 本模块导入路径, Releases 仓库 (owner/repo) 与 go install 委托均由此得出
	module = "github.com/wujunwei928/dev"

	// defaultAPIBase GitHub API 默认 base, 可注入覆盖 (唯一为可测试性设的注入点)
	defaultAPIBase = "https://api.github.com"
)

// CheckResult 版本检查结果
type CheckResult struct {
	Current  string // 当前版本展示串
	Latest   string // 最新可用版本, 如 "v0.2.1"
	UpToDate bool   // 当前已是最新
}

// Check 只读查询最新可用版本并与当前版本比较, 不执行任何更新动作。
// apiBase 为 GitHub API base, 传空串时使用默认 https://api.github.com。
func Check(current VersionInfo, apiBase string) (CheckResult, error) {
	latest, err := latestVersion(current.Method, apiBase)
	if err != nil {
		return CheckResult{}, err
	}
	return CheckResult{
		Current:  current.Version,
		Latest:   latest,
		UpToDate: normalizeVersion(current.Version) == normalizeVersion(latest),
	}, nil
}

// latestVersion 按安装方式选择查询通道:
// go install 构建经 go list -m (与更新通道共享 GOPROXY 语义), 其余查 GitHub Releases API
func latestVersion(m Method, apiBase string) (string, error) {
	if m == MethodGoInstall {
		return latestViaGo()
	}
	return latestViaAPI(apiBase)
}

// latestViaGo 委托 go 命令查询模块最新发布版本
func latestViaGo() (string, error) {
	// #nosec G204 -- go 命令与模块路径均为常量, 无用户输入
	out, err := exec.Command("go", "list", "-m", module+"@latest").Output()
	if err != nil {
		stderr := ""
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		}
		if stderr != "" {
			return "", fmt.Errorf("查询最新版本失败: 执行 go list -m: %v: %s", err, stderr)
		}
		return "", fmt.Errorf("查询最新版本失败: 执行 go list -m: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) < 2 {
		return "", fmt.Errorf("查询最新版本失败: go list -m 输出异常: %q", strings.TrimSpace(string(out)))
	}
	return fields[1], nil
}

// repoSlug 从模块路径提取 GitHub owner/repo ("github.com/wujunwei928/dev" -> "wujunwei928/dev")
func repoSlug() string {
	parts := strings.Split(module, "/")
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

// latestViaAPI 查询 GitHub Releases 的 latest release tag, 复用 fetchRelease 的请求与解析
func latestViaAPI(apiBase string) (string, error) {
	rel, err := fetchRelease(apiBase, "")
	if err != nil {
		return "", err
	}
	return rel.TagName, nil
}

// normalizeVersion 去掉前导 v/V, 用于版本号等值比较 ("v0.2.1" 与 "0.2.1" 等价)
func normalizeVersion(s string) string {
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		return s[1:]
	}
	return s
}
