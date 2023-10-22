package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

type timeSubCmd struct{}

func NewCmdTime() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "time",
		Short: "时间格式处理",
		Long:  `时间格式处理`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// 添加子命令
	timeCmd := timeSubCmd{}
	cmd.AddCommand(timeCmd.NewCmdParse())

	return cmd
}

func (c timeSubCmd) NewCmdParse() *cobra.Command {
	// 解析时间抽参数
	var parseTimeStamp int64

	cmd := &cobra.Command{
		Use:   "parse",
		Short: "解析时间戳为标准格式, 不传默认当前时间",
		Long:  "解析时间戳为标准格式, 不传默认当前时间",
		Run: func(cmd *cobra.Command, args []string) {
			nowTime := time.Unix(parseTimeStamp, 0)
			log.Printf("当前时间: %s , %d", nowTime.Format("2006-01-02 15:04:05"), nowTime.Unix())
		},
	}

	cmd.Flags().Int64VarP(&parseTimeStamp, "timestamp", "t", time.Now().Unix(), "解析时间戳, 不传默认当前时间戳")

	return cmd
}
