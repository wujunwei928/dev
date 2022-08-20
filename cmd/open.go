package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/search"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "打开网址或文件路径",
	Long:  `打开网址或文件路径, 网址需要协议, 如:https://`,
	Args:  cobra.ExactArgs(1), // 只支持一个args
	Run: func(cmd *cobra.Command, args []string) {
		searchStr := args[0]
		if strings.Index(searchStr, "~") == 0 {
			// 以~开头时, Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				log.Fatalf("get home dir fail: %s", err.Error())
			}
			searchStr = strings.ReplaceAll(searchStr, "~", home)
		}
		fmt.Println(searchStr)
		search.Open(searchStr)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
