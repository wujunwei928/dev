package cmd

import (
	"log"
	"strings"

	"github.com/wujunwei928/dev/internal/word"

	"github.com/spf13/cobra"
)

const (
	ModeUpper                      = iota + 1 // 全部转大写
	ModeLower                                 // 全部转小写
	ModeUnderscoreToUpperCamelCase            // 下划线转大写驼峰
	ModeUnderscoreToLowerCamelCase            // 下线线转小写驼峰
	ModeCamelCaseToUnderscore                 // 驼峰转下划线
)

var desc = strings.Join([]string{
	"该子命令支持各种单词格式转换，模式如下：",
	"1：全部转大写",
	"2：全部转小写",
	"3：下划线转大写驼峰",
	"4：下划线转小写驼峰",
	"5：驼峰转下划线",
}, "\n")

var wordStr string
var wordMode int8

// wordCmd represents the word command
var wordCmd = &cobra.Command{
	Use:   "word",
	Short: "单词格式转换",
	Long:  desc,
	Run: func(cmd *cobra.Command, args []string) {
		var content string
		switch wordMode {
		case ModeUpper:
			content = word.ToUpper(wordStr)
		case ModeLower:
			content = word.ToLower(wordStr)
		case ModeUnderscoreToUpperCamelCase:
			content = word.UnderscoreToUpperCamelCase(wordStr)
		case ModeUnderscoreToLowerCamelCase:
			content = word.UnderscoreToLowerCamelCase(wordStr)
		case ModeCamelCaseToUnderscore:
			content = word.CamelCaseToUnderscore(wordStr)
		default:
			log.Fatalf("暂不支持该转换模式, 请执行 help word 查看帮助文档")
		}

		log.Printf("输出结果: %s", content)
	},
}

func init() {
	rootCmd.AddCommand(wordCmd)

	wordCmd.Flags().StringVarP(&wordStr, "str", "s", "", "请输入单词内容")
	wordCmd.Flags().Int8VarP(&wordMode, "mode", "m", 0, "请输入单词转化的模式")
}
