package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wujunwei928/rd/internal/search"
)

var searchStr string
var searchMode string
var searchType string

var searchTypeUSage = strings.Join([]string{
	"检索方式: ",
	"browser: 打开默认浏览器检索",
	"cli: 终端显示搜索内容",
}, "\n")

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "搜索",
	Long:  "指定搜索引擎, 检索相关query",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(searchMode) <= 0 {
			searchMode = viper.GetString("default_search_engine")
		}
		switch searchType {
		case "cli":
			searchRes, _ := search.RequestDetail(searchMode, searchStr)
			keyStyle := pterm.NewStyle(pterm.FgLightBlue, pterm.Bold) // 标题cli样式
			for i, s := range searchRes {
				ptermTable := pterm.TableData{
					{keyStyle.Sprint("序号"), strconv.Itoa(i + 1)},
				}
				for _, v := range s {
					ptermTable = append(ptermTable, []string{keyStyle.Sprint(v.Key), v.Val})
				}
				pterm.DefaultTable.
					WithHasHeader(false).
					WithData(ptermTable).
					WithLeftAlignment().
					Render()
				fmt.Print("\n\n\n")
			}
		default:
			searchUrl := search.FormatSearchUrl(searchMode, searchStr)
			err = search.Open(searchUrl)

		}
		if err != nil {
			log.Fatalf("检索失败: %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchStr, "str", "s", "", "请输入搜索query")
	searchCmd.Flags().StringVarP(&searchMode, "mode", "m", "", search.FormatCommandDesc())
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", searchTypeUSage)
}
