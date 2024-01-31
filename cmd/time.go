package cmd

import (
	"log"
	"strconv"
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
	subCmd := timeSubCmd{}
	cmd.AddCommand(subCmd.NewCmdParse())

	return cmd
}

func (c timeSubCmd) NewCmdParse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "解析时间戳为标准格式, 不传默认当前时间",
		Long:  "解析时间戳为标准格式, 不传默认当前时间",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			parseTimeStamp, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatalf("strconv.ParseInt err: %v", err)
			}
			nowTime := time.Unix(parseTimeStamp, 0)
			log.Printf("当前时间: %s , %d", nowTime.Format("2006-01-02 15:04:05"), nowTime.Unix())
		},
	}

	return cmd
}
