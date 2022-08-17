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
var searchCliDesc bool

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
			// 终端显示搜索结果
			searchRes, err := search.RequestDetail(searchMode, searchStr)
			if err != nil {
				log.Fatalf("request search engine fail: %s", err.Error())
			}
			keyStyle := pterm.NewStyle(pterm.FgLightBlue, pterm.Bold) // 标题cli样式
			termRenderList := make([]string, 0, len(searchRes))
			for i, s := range searchRes {
				ptermTable := pterm.TableData{
					{keyStyle.Sprint("序号"), strconv.Itoa(i + 1)},
				}
				for _, v := range s {
					ptermTable = append(ptermTable, []string{keyStyle.Sprint(v.Key), v.Val})
				}
				itemRender, err := pterm.DefaultTable.
					WithHasHeader(false).
					WithData(ptermTable).
					WithLeftAlignment().
					Srender()
				if err != nil {
					continue
				}

				// 根据显示顺序, 判断切片追加方向
				if searchCliDesc {
					termRenderList = append([]string{itemRender}, termRenderList...)
				} else {
					termRenderList = append(termRenderList, itemRender)
				}
			}
			// 打印终端显示
			fmt.Println(strings.Join(termRenderList, "\n\n\n"))
		default:
			// 打开默认浏览器搜索
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
	searchCmd.Flags().BoolVarP(&searchCliDesc, "desc", "", true, "是否倒序展示: 默认倒序, 方便查看(只终端展示生效)")
}
