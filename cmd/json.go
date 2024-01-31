package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/json2struct"
)

type jsonSubCmd struct{}

func NewCmdJson() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json",
		Short: "json转换和处理",
		Long:  "json转换和处理",
		//Run:   func(cmd *cobra.Command, args []string) {}, // 这里如果没有处理逻辑的话, 可以不设置, 否则不会显示提示信息
	}

	// 添加子命令
	jsonCmd := jsonSubCmd{}
	cmd.AddCommand(jsonCmd.NewCmdStruct())
	cmd.AddCommand(jsonCmd.NewCmdPrint())

	return cmd
}

func (j jsonSubCmd) NewCmdStruct() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "struct",
		Short: "json转换golang结构体",
		Long:  `json转换golang结构体, json需要单引号包裹, cmd下不支持, 需要转义`,
		Example: strings.Join([]string{
			`非windows cmd.exe终端, json单引号包裹: dev json struct '{"name":"wjw","age":123}'`,
			`windows cmd终端, 单引号不支持, json需转义: dev json struct {\"name\":\"wjw\",\"age\":123}`,
		}, "\n"),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jsonStr := args[0]
			parser, err := json2struct.NewParser(jsonStr)
			if err != nil {
				log.Fatalf("json2struct.NewParser err: %v", err)
			}
			content := parser.Json2Struct()
			log.Printf("输出结果: \n%s", content)
		},
	}

	return cmd
}

func (j jsonSubCmd) NewCmdPrint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "格式化json，类似jq",
		Long:  "格式化json，类似jq",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jsonStr := args[0]
			var jsonInterface interface{}
			err := json.Unmarshal([]byte(jsonStr), &jsonInterface)
			if err != nil {
				log.Fatalln(pterm.Red("输入非标准json，报错信息：", err.Error()))
			}
			jsonBytes, err := json.MarshalIndent(jsonInterface, "", "    ")
			if err != nil {
				log.Fatalln(pterm.Red("json格式化失败，报错信息：", err.Error()))
			}
			fmt.Println(pterm.Green(string(jsonBytes)))
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdJson())
}
