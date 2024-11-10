package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wujunwei928/dev/internal/word"
)

const (
	ModeUpper                      = iota + 1 // 全部转大写
	ModeLower                                 // 全部转小写
	ModeUnderscoreToUpperCamelCase            // 下划线转大写驼峰
	ModeUnderscoreToLowerCamelCase            // 下线线转小写驼峰
	ModeCamelCaseToUnderscore                 // 驼峰转下划线
)

func NewCmdWord() *cobra.Command {
	var (
		wordMode int8
	)

	cmd := &cobra.Command{
		Use:   "word",
		Short: "单词格式转换",
		Long:  "单词格式转换",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			wordStr := args[0]

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

	cmd.Flags().Int8VarP(&wordMode, "mode", "m", 0, strings.Join([]string{
		"单词转换模式, 支持模式如下: ",
		"1：全部转大写",
		"2：全部转小写",
		"3：下划线转大写驼峰",
		"4：下划线转小写驼峰",
		"5：驼峰转下划线",
	}, "\n"))

	return cmd
}
