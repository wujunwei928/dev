package updater

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// smokeTestTimeout 冒烟测试单次执行上限, 防止异常产物挂起更新流程
const smokeTestTimeout = 30 * time.Second

// smokeTest 对暂存位置的新二进制执行 --version 冒烟测试:
// 输出中的版本号必须与目标版本一致, 否则拒绝替换 (坏产物永不就位)
func smokeTest(binaryPath, target string) error {
	ctx, cancel := context.WithTimeout(context.Background(), smokeTestTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, binaryPath, "--version").Output()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			detail += " stderr: " + strings.TrimSpace(string(exitErr.Stderr))
		}
		return fmt.Errorf("冒烟测试失败: 新二进制 --version 执行出错: %v (%s), 已拒绝替换", err, detail)
	}
	if !outputMatchesVersion(string(out), target) {
		return fmt.Errorf("冒烟测试失败: 新二进制 --version 输出 %q, 与目标版本 %s 不符, 已拒绝替换",
			strings.TrimSpace(string(out)), target)
	}
	return nil
}

// outputMatchesVersion 判断 --version 输出是否包含与目标一致的版本号:
// 任一空白分隔 token 经 v 前缀归一后相等即匹配 (cobra 默认输出形如 "dev version 0.2.1")
func outputMatchesVersion(output, target string) bool {
	want := normalizeVersion(strings.TrimSpace(target))
	for _, f := range strings.Fields(output) {
		if normalizeVersion(f) == want {
			return true
		}
	}
	return false
}
