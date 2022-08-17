package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/rd/internal/json2struct"
)

var jsonStr string

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "json转换和处理",
	Long:  "json转换和处理",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var json2structCmd = &cobra.Command{
	Use:   "struct",
	Short: "json转换  ps:json双引号需要转义",
	Long: `json转换 ps:json双引号需要转义
	example: go run main.go json struct -s={\"name\":\"wjw\",\"age\":123}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println(jsonStr, len(jsonStr), args, os.Args)
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
	//json2structCmd.Flags().StringVarP(&jsonStr, "str", "s", "", "请输入json字符串")
	json2structCmd.PersistentFlags().StringVarP(&jsonStr, "str", "s", "", "请输入json字符串")
}
