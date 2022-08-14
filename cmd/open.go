package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wujunwei928/bd/internal/search"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "打开网址或文件路径",
	Long:  `打开网址或文件路径, 网址需要协议如:https://`,
	Run: func(cmd *cobra.Command, args []string) {
		search.Open(searchStr)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().StringVarP(&searchStr, "str", "s", "", "请输入搜索query")
}
