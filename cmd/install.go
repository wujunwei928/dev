package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
)

func NewCmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "安装命令到PATH",
		Long:  `安装命令到PATH, 方便使用`,
		Run: func(cmd *cobra.Command, args []string) {
			// 获取环境变量的文件夹列表
			pathList := parseEnvPath()
			if len(pathList) <= 0 {
				log.Fatalln("no path detected, you can manually install dev by copying the binary to path folder.")
			}

			// 终端选择项
			var options []string
			for _, s := range pathList {
				options = append(options, fmt.Sprintf("%s", s))
			}
			selectedOption, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("请选择要安装的路径: [可上方向键移动, 或输入关键字符检索, 回车选择]").
				WithOptions(options).
				WithMaxHeight(6).
				Show()
			if len(selectedOption) <= 0 {
				log.Fatalln("selected path option is empty, please check")
			}
			pterm.Info.Printfln("selected path option: %s", pterm.Green(selectedOption))

			// 复制文件到选择的PATH文件夹
			// 命令行参数解析命令名(go run命令运行, 也会生成一个临时可执行文件)
			copyFromFile, err := filepath.Abs(os.Args[0])
			if err != nil {
				pterm.Warning.Printf(err.Error())
				log.Fatalln("install fail, you can manually install dev by copying the binary to path folder.")
			}
			// window cmd在当前文件夹运行时, 可以不带exe
			if runtime.GOOS == "windows" && !strings.Contains(copyFromFile, ".exe") {
				copyFromFile = copyFromFile + ".exe"
			}
			// 需要复制到选择PATH项的文件名
			copyToFile, err := filepath.Abs(filepath.Join(selectedOption, filepath.Base(copyFromFile)))
			pterm.Info.Printfln("copy from: " + copyFromFile)
			pterm.Info.Printfln("copy to: " + copyToFile)
			if err != nil {
				pterm.Warning.Printf(err.Error())
				log.Fatalln("install fail, you can manually install dev by copying the binary to path folder.")
			}
			// 读取当前运行文件内容
			input, err := os.ReadFile(copyFromFile)
			if err != nil {
				pterm.Warning.Printf(err.Error())
				log.Fatalln("install fail, you can manually install dev by copying the binary to path folder.")
			}
			// 将文件内容写到目标文件
			err = os.WriteFile(copyToFile, input, 0744)
			if err != nil {
				pterm.Warning.Printf(err.Error())
				log.Fatalln("install fail, you can manually install dev by copying the binary to path folder.")
			}
			pterm.Success.Println("install success")
		},
	}

	return cmd
}

func parseEnvPath() []string {
	envPath := os.Getenv("PATH")
	if len(envPath) <= 0 {
		return nil
	}
	pathList := strings.Split(envPath, ":")
	if runtime.GOOS == "windows" {
		pathList = strings.Split(envPath, ";")
	}

	formatList := make([]string, 0, len(pathList))
	// 优先GOPATH 和 GOROOT
	goPathBin := getGoPathBin()
	if len(goPathBin) > 0 && strings.Contains(envPath, goPathBin) {
		formatList = append(formatList, goPathBin)
	}
	goRootBin := getGoRootBin()
	if len(goRootBin) > 0 && strings.Contains(envPath, goRootBin) {
		formatList = append(formatList, goRootBin)
	}
	// other path
	for _, s := range pathList {
		if len(s) > 0 && s != goPathBin && s != goRootBin {
			formatList = append(formatList, s)
		}
	}

	return formatList
}

func getGoPathBin() string {
	envGoPath := os.Getenv("GOPATH")
	if len(envGoPath) <= 0 {
		return ""
	}
	goPathBin, err := filepath.Abs(path.Join(envGoPath, "bin"))
	if err != nil {
		return ""
	}
	return goPathBin
}

func getGoRootBin() string {
	envGoRoot := os.Getenv("GOROOT")
	if len(envGoRoot) <= 0 {
		return ""
	}
	goRootBin, err := filepath.Abs(path.Join(envGoRoot, "bin"))
	if err != nil {
		return ""
	}
	return goRootBin
}
