package cmd

import (
	"errors"
	"fmt"
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
	SearchConfigEngine     = "search.default_engine"
	SearchConfigType       = "search.default_type"
	SearchConfigCliIsDesc  = "search.cli_is_desc"
	SearchConfigConcurrent = "search.concurrent"
)

func NewCmdSearch() *cobra.Command {
	var concurrent bool
	var jsonOutput bool
	var searchTypeUSage = strings.Join([]string{
		"检索方式: ",
		"cli: 终端显示搜索内容",
		"browser: 打开默认浏览器检索",
	}, "\n")

	cmd := &cobra.Command{
		Use:     "search",
		Aliases: []string{"??"},
		Short:   "搜索",
		Long:    "指定搜索引擎, 检索相关query",
		Example: getSearchExample(),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New(pterm.Red("requires at least 1 arg(s), only received 0"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var searchRes [][]search.KeyVal
			if concurrent {
				// 并发搜索多个主流搜索引擎
				searchEngines := []string{search.EngineBing, search.EngineBaidu, search.EngineGoogle}
				concurrentResults := search.ConcurrentSearch(searchEngines, searchStr)
				for _, cr := range concurrentResults {
					if cr.Err != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "搜索引擎 %s 请求失败: %v\n", cr.Engine, cr.Err)
						continue
					}
					searchRes = append(searchRes, cr.Items...)
				}
			} else {
				// 单个搜索引擎搜索
				searchRes, err = search.RequestDetail(searchMode, searchStr)
				if err != nil {
					return fmt.Errorf("request search engine fail: %w", err)
				}
			}

			switch searchType {
			case SearchTypeCli:
				// JSON 模式输出
				if jsonOutput {
					result := search.KeyValToResultItems(searchMode, searchStr, searchRes)
					jsonStr, err := result.ToJSON()
					if err != nil {
						return fmt.Errorf("JSON 序列化失败: %w", err)
					}
					fmt.Println(jsonStr)
					return nil
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
				return fmt.Errorf("检索失败: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringP("mode", "m", DefaultSearchEngine, search.FormatSearchCommandModeUsage())
	cmd.Flags().StringP("type", "t", DefaultSearchType, searchTypeUSage)
	cmd.Flags().BoolP("desc", "", DefaultCliIsDesc, "终端是否倒序展示: 默认倒序, 方便查看")
	cmd.Flags().BoolVar(&concurrent, "concurrent", false, "并发搜索多个搜索引擎 (Bing, Baidu, Google)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "以 JSON 格式输出搜索结果")

	// flag 和 viper 绑定
	viper.BindPFlag(SearchConfigEngine, cmd.Flags().Lookup("mode"))
	viper.BindPFlag(SearchConfigType, cmd.Flags().Lookup("type"))
	viper.BindPFlag(SearchConfigCliIsDesc, cmd.Flags().Lookup("desc"))
	viper.BindPFlag(SearchConfigConcurrent, cmd.Flags().Lookup("concurrent"))

	return cmd
}

// search命令使用example
func getSearchExample() string {
	pTermTable := pterm.TableData{
		{"描述", "命令"},
		{`默认配置检索`, `dev search golang`},
		{`指定搜索引擎检索`, `dev search -m baidu -t cli "docker practice"`},
		{`打开系统默认浏览器检索`, `dev search -m baidu -t browser "k8s"`},
		{`终端正序显示搜索结果`, `dev search -m bing -t cli --desc=false "golang cobra"`},
		{`常规搜索引擎site检索`, `dev search -m bing "golang site:cnblogs.com"`},
		{`JSON 格式输出搜索结果`, `dev search -m bing -t cli --json "golang orm"`},
	}
	itemRender, err := pterm.DefaultTable.
		WithData(pTermTable).
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
