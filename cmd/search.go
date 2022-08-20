package cmd

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wujunwei928/dev/internal/search"
)

// 检索类型
const (
	SearchTypeBrowser = "browser" // 浏览器
	SearchTypeCli     = "cli"     // 终端显示
)

// 搜索默认项
const (
	DefaultSearchEngine = search.EngineBing // 默认搜索引擎
	DefaultSearchType   = SearchTypeCli     // 默认检索类型
	DefaultCliIsDesc    = true              // 默认cli是否倒序显示
)

// 搜索配置项
const (
	SearchConfigEngine    = "search.default_engine"
	SearchConfigType      = "search.default_type"
	SearchConfigCliIsDesc = "search.cli_is_desc"
)

var searchTypeUSage = strings.Join([]string{
	"检索方式: ",
	"cli: 终端显示搜索内容",
	"browser: 打开默认浏览器检索",
}, "\n")

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "搜索",
	Long:    "指定搜索引擎, 检索相关query",
	Example: getSearchExample(),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New(pterm.Red("requires at least 1 arg(s), only received 0"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		searchStr := strings.Join(args, " ") // 多个args以空格隔开
		//fmt.Println(searchStr)
		searchMode := viper.GetString(SearchConfigEngine)
		searchType := viper.GetString(SearchConfigType)
		searchCliDesc := viper.GetBool(SearchConfigCliIsDesc)
		//fmt.Println(searchMode, searchType, searchCliDesc)
		if len(searchMode) <= 0 {
			searchMode = viper.GetString("default_search_engine")
		}
		switch searchType {
		case SearchTypeCli:
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

	searchCmd.Flags().StringP("mode", "m", DefaultSearchEngine, search.FormatSearchCommandModeUsage())
	searchCmd.Flags().StringP("type", "t", DefaultSearchType, searchTypeUSage)
	searchCmd.Flags().BoolP("desc", "", DefaultCliIsDesc, "终端是否倒序展示: 默认倒序, 方便查看")

	// flag 和 viper 绑定
	viper.BindPFlag(SearchConfigEngine, searchCmd.Flags().Lookup("mode"))
	viper.BindPFlag(SearchConfigType, searchCmd.Flags().Lookup("type"))
	viper.BindPFlag(SearchConfigCliIsDesc, searchCmd.Flags().Lookup("desc"))
}

// search命令使用example
func getSearchExample() string {
	ptermTable := pterm.TableData{
		{"描述", "命令"},
		{`默认配置检索`, `dev search golang`},
		{`指定搜索引擎检索`, `dev search -m baidu -t cli "docker practice"`},
		{`打开系统默认浏览器检索`, `dev search -m baidu -t browser "k8s"`},
		{`终端正序显示搜索结果`, `dev search -m bing -t cli --desc=false "golang cobra"`},
		{`常规搜索引擎site检索`, `dev search -m bing "golang site:cnblogs.com"`},
	}
	itemRender, err := pterm.DefaultTable.
		WithData(ptermTable).
		WithHasHeader(true).
		WithHeaderRowSeparator("-").
		WithBoxed(true).
		WithRowSeparator("-").
		WithLeftAlignment().
		Srender()
	if err != nil {
		return ""
	}
	return itemRender
}
