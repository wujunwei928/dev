package cmd

import (
	"log"

	"github.com/wujunwei928/bd/internal/search"

	"github.com/spf13/cobra"
)

var searchStr string
var searchMode string

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "搜索",
	Long:  search.FormatCommandDesc(),
	Run: func(cmd *cobra.Command, args []string) {
		searchUrl := search.FormatSearchUrl(searchMode, searchStr)
		err := search.Open(searchUrl)
		if err != nil {
			log.Fatalf("检索失败: %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchStr, "str", "s", "", "请输入搜索query")
	searchCmd.Flags().StringVarP(&searchMode, "mode", "m", "", "请输入搜索引擎")
}
