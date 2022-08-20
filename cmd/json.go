package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/json2struct"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "json转换和处理",
	Long:  "json转换和处理",
	//Run:   func(cmd *cobra.Command, args []string) {}, // 这里如果没有处理逻辑的话, 可以不设置, 否则不会显示提示信息
}

var json2structCmd = &cobra.Command{
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

func init() {
	rootCmd.AddCommand(jsonCmd)

	jsonCmd.AddCommand(json2structCmd)
}
