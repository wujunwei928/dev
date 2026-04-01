package cmd

import (
	"os/exec"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/wujunwei928/dev/internal/search"
)

var ConsolePromptSuggestList = []prompt.Suggest{
	{Text: "help", Description: "查看帮助，列出所有命令"},
	{Text: "//", Description: "使用默认浏览器打开网址"},
	{Text: "??", Description: "使用搜索引擎搜索关键字"},
	{Text: ">", Description: "执行命令行命令"},
	{Text: "open", Description: "使用默认程序打开文件"},
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
			historyList := make([]string, 0, 100)
			for {
				t := prompt.Input("> ", consoleCompleter, prompt.OptionHistory(historyList))
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

				isValidCommand := true
				keyWord := strings.Join(consoleArgs[1:], " ")
				switch consoleArgs[0] {
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
					isValidCommand = false
					println("暂不支持该命令")
				}
				if isValidCommand {
					historyList = append(historyList, t)
				}
			}
		},
	}

	return cmd
}
