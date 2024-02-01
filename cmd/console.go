package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pterm/pterm"

	"github.com/wujunwei928/dev/internal/search"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var ConsolePromptSuggestList = []prompt.Suggest{
	{Text: "help", Description: "查看帮助，列出所有命令"},
	{Text: "//", Description: "使用默认浏览器打开网址"},
	{Text: "??", Description: "使用搜索引擎搜索关键字"},
	{Text: ">", Description: "执行命令行命令"},
	{Text: "open", Description: "使用默认程序打开文件"},
	{Text: "md5", Description: "md5加密"},
	{Text: "sha1", Description: "sha1加密"},
	{Text: "base64_encode", Description: "base64加密"},
	{Text: "base64_decode", Description: "base64解密"},
	{Text: "url_encode", Description: "url加密"},
	{Text: "url_decode", Description: "url解密"},
	{Text: "unicode_encode", Description: "unicode加密"},
	{Text: "unicode_decode", Description: "unicode解密"},
}

func consoleCompleter(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(ConsolePromptSuggestList, d.GetWordBeforeCursor(), true)
}

func showConsoleHelp() {
	println("【 空行使用Tab键也可触发提示可用命令 】")
	pTermTable := pterm.TableData{
		{"命令", "描述"},
	}
	for _, v := range ConsolePromptSuggestList {
		pTermTable = append(pTermTable, []string{pterm.Green(v.Text), v.Description})
	}
	pterm.DefaultTable.WithData(pTermTable).WithHasHeader(true).Render()
}

func NewCmdConsole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "console",
		Short: "类似ipython的交互式命令行",
		Long:  "类似ipython的交互式命令行",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				t := prompt.Input("> ", consoleCompleter)
				t = strings.TrimSpace(t)
				if t == "exit" {
					break
				}
				consoleArgs := strings.Split(t, " ")
				if len(consoleArgs) <= 0 {
					continue
				}
				if consoleArgs[0] == "help" {
					showConsoleHelp()
					continue
				}
				if len(consoleArgs) < 2 {
					println("参数不足")
					continue
				}
				keyWord := strings.Join(consoleArgs[1:], " ")
				switch consoleArgs[0] {
				case "md5":
					hash := md5.Sum([]byte(keyWord))
					content := hex.EncodeToString(hash[:])
					println("md5加密结果: " + content)
				case "sha1":
					hash := sha1.New()
					hash.Write([]byte(keyWord))
					content := hex.EncodeToString(hash.Sum(nil))
					println("sha1加密结果: " + content)
				case "base64_encode":
					content := base64.StdEncoding.EncodeToString([]byte(keyWord))
					println("base64加密结果: " + content)
				case "base64_decode":
					content, err := base64.StdEncoding.DecodeString(keyWord)
					if err != nil {
						println("base64解密失败: " + err.Error())
						continue
					}
					println("base64解密结果: " + string(content))
				case "url_encode":
					content := url.QueryEscape(keyWord)
					println("url加密结果: " + content)
				case "url_decode":
					content, err := url.QueryUnescape(keyWord)
					if err != nil {
						println("url解密失败: " + err.Error())
						continue
					}
					println("url解密结果: " + content)
				case "unicode_encode":
					content := ""
					for _, v := range keyWord {
						content += fmt.Sprintf("\\u%04x", v)
					}
					println("unicode加密结果: " + content)
				case "unicode_decode":
					// unicode转换中文需要的格式 "内容" , 注意要传双引号
					if !strings.Contains(keyWord, `"`) {
						keyWord = `"` + keyWord + `"`
					}
					s, err := strconv.Unquote(keyWord)
					if err != nil {
						println("unicode解密失败: " + err.Error())
						continue
					}
					println("unicode解密结果: " + s)
				case "open":
					err := search.Open(keyWord)
					if err != nil {
						println("打开文件失败: " + err.Error())
					}
				case "//":
					openUrl := "https://" + strings.TrimSpace(keyWord)
					err := search.Open(openUrl)
					if err != nil {
						println("打开网址失败: " + err.Error())
					}
				case "??":
					searchUrl := search.FormatSearchUrl(search.EngineBing, keyWord)
					err := search.Open(searchUrl)
					if err != nil {
						println("搜索失败: " + err.Error())
					}
				case ">":
					runName := consoleArgs[1]
					runArgs := make([]string, 0, 1)
					if len(consoleArgs) > 2 {
						runArgs = consoleArgs[2:]
					}
					runCmd := exec.Command(runName, runArgs...)
					runCmdOutput, runCmdErr := runCmd.CombinedOutput()
					if runCmdErr != nil {
						println("执行命令失败: " + runCmdErr.Error() + ", " + string(runCmdOutput))
						continue
					}
					println(string(runCmdOutput))
				default:
					println("暂不支持该命令")
				}
			}
		},
	}

	return cmd
}
