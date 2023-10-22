package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/search"
)

func NewCmdOpen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "打开网址或文件路径",
		Long:  `打开网址或文件路径, 网址需要协议, 如:https://`,
		Args:  cobra.ExactArgs(1), // 只支持一个args
		Run: func(cmd *cobra.Command, args []string) {
			searchStr := args[0]
			fmt.Println(searchStr)
			search.Open(searchStr)
		},
	}

	return cmd
}
