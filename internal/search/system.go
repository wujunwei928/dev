package search

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// SystemCallCommands 系统调用命令
// 参考: GOROOT/src/cmd/internal/browser/browser.go
var SystemCallCommands = map[string][]string{
	"windows": {"cmd", "/c", "start"},
	"darwin":  {"open"},
	"linux":   {"xdg-open"},
}

// Open 调用系统命令打开网址或文件夹
func Open(path string) error {
	cmd, ok := SystemCallCommands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	// 判断是否在windows wsl环境中
	if runtime.GOOS == "linux" {
		unameCmd := exec.Command("uname", "-a")
		stdoutStdErr, err := unameCmd.CombinedOutput()
		if err != nil {
			log.Fatalln("check is wsl fail" + err.Error())
		}
		if strings.Contains(strings.ToLower(string(stdoutStdErr)), "microsoft") {
			cmd = []string{"explorer.exe"}
		}
	}
	// 终端执行命令, 如果网址包含&符号, 需要进行转义
	if strings.Contains(path, "&") {
		switch runtime.GOOS {
		case "windows":
			// windows将&前添加^
			path = strings.ReplaceAll(path, "&", "^&")
		default:
			// linux,mac&前添加\
			path = strings.ReplaceAll(path, "&", `\&`)
		}
	}
	cmd = append(cmd, path)

	command := exec.Command(cmd[0], cmd[1:]...)
	return command.Start()
}
