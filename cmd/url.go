package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/search"
)

func NewCmdUrl() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "url",
		Aliases: []string{"//"},
		Short:   "打开网址",
		Long:    `打开网址`,
		Args:    cobra.ExactArgs(1), // 只支持一个args
		Run: func(cmd *cobra.Command, args []string) {
			searchStr := args[0]
			fmt.Println(searchStr)
			if !strings.HasPrefix(searchStr, "http") {
				searchStr = "https://" + searchStr
			}
			search.Open(searchStr)
		},
	}

	return cmd
}
