package search

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemCallCommands 系统调用命令
// windows 使用 rundll32 打开: 不经过 cmd.exe 元字符重解释,
// 含 & | % 等字符的网址和带空格的本地路径可原样安全传递
var SystemCallCommands = map[string][]string{
	"windows": {"rundll32", "url.dll,FileProtocolHandler"},
	"darwin":  {"open"},
	"linux":   {"xdg-open"},
}

// wslOpenCommand wsl环境下打开网址/路径的命令
var wslOpenCommand = []string{"explorer.exe"}

// Open 调用系统命令打开网址或文件夹
func Open(path string) error {
	if strings.Index(path, "~") == 0 {
		// 以~开头时, Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir fail: %w", err)
		}
		path = strings.ReplaceAll(path, "~", home)
	}

	base, ok := SystemCallCommands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	// 判断是否在windows wsl环境中
	if runtime.GOOS == "linux" {
		isWsl, err := isWslEnvironment()
		if err != nil {
			return err
		}
		if isWsl {
			base = wslOpenCommand
		}
	}

	// exec.Command 不经过shell, 路径中的特殊字符无需转义
	args := appendOpenArgs(base, path)
	// #nosec G204 -- 命令来自固定平台映射表(rundll32/open/xdg-open/explorer), 不经过shell
	command := exec.Command(args[0], args[1:]...)
	return command.Start()
}

// buildOpenArgs 按平台构造打开路径的完整命令参数, 未知平台返回nil
func buildOpenArgs(goos, path string) []string {
	base, ok := SystemCallCommands[goos]
	if !ok {
		return nil
	}
	return appendOpenArgs(base, path)
}

// appendOpenArgs 在基础命令后拼接路径, 不修改原切片
func appendOpenArgs(base []string, path string) []string {
	args := make([]string, 0, len(base)+1)
	args = append(args, base...)
	args = append(args, path)
	return args
}

// isWslEnvironment 检测当前linux环境是否为wsl
func isWslEnvironment() (bool, error) {
	unameCmd := exec.Command("uname", "-a")
	stdoutStdErr, err := unameCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("check is wsl fail: %w", err)
	}
	return strings.Contains(strings.ToLower(string(stdoutStdErr)), "microsoft"), nil
}
