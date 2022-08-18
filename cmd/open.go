package cmd

import (
	"log"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/search"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "打开网址或文件路径",
	Long:  `打开网址或文件路径, 网址需要协议如:https://`,
	Run: func(cmd *cobra.Command, args []string) {
		if searchStr == "~" {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				log.Fatalf("get home dir fail: %s", err.Error())
			}
			searchStr = home
		}
		search.Open(searchStr)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().StringVarP(&searchStr, "str", "s", "", "请输入搜索query")
}
