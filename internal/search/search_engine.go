package search

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// KeyVal Golang中map遍历没有顺序, 使用[]KeyVal代替map[string]string返回
type KeyVal struct {
	Key string
	Val string
}

func RequestDetail(searchEngine string, query string) ([][]KeyVal, error) {
	var (
		parseRes [][]KeyVal
		err      error
	)

	engineParam := getEngineParam(searchEngine)

	reqUrl := engineParam.Domain + strings.ReplaceAll(engineParam.Param, "{search_query}", url.QueryEscape(query))
	referUrl := engineParam.Domain
	if len(engineParam.AjaxUrl) > 0 {
		// 是否是ajax请求
		referUrl = reqUrl
		reqUrl = strings.ReplaceAll(engineParam.AjaxUrl, "{search_query}", url.QueryEscape(query))
	}

	client := resty.New()
	client.SetTimeout(2000 * time.Millisecond)
	res, err := client.R().
		EnableTrace().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36 Edg/104.0.1293.54").
		SetHeader("Referer", referUrl).
		Get(reqUrl)
	if err != nil {
		return nil, err
	}
	//fmt.Println(reqUrl, string(res.Body()), res.StatusCode())

	if engineParam.JsonRule != nil {
		parseRes, err = engineParam.parseJson(res.Body())
	} else if engineParam.HtmlRule != nil {
		parseRes, err = engineParam.parseHtml(res.Body())
	} else {
		return nil, errors.New("json parse rule and html parse rule is both nil, please check")
	}

	return parseRes, nil
}

// 解析html结果
func (e EngineParam) parseHtml(httpBody []byte) ([][]KeyVal, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(httpBody))
	if err != nil {
		return nil, err
	}

	parseRes := make([][]KeyVal, 0, 10)
	//fmt.Println(engineParam.HtmlRule.ListRule)
	doc.Find(e.HtmlRule.ListRule).Each(func(i int, s *goquery.Selection) {
		item := make([]KeyVal, 0, len(e.HtmlRule.ListItemRule))
		for _, rule := range e.HtmlRule.ListItemRule {
			var itemContent string
			if len(rule.Attr) > 0 {
				itemContent, _ = s.Find(rule.Rule).Attr(rule.Attr)
				if rule.Attr == "href" &&
					!strings.Contains(itemContent, "http://") &&
					!strings.Contains(itemContent, "https://") {
					// 有些网站(如: 搜狗)的跳转链接不带域名, 需要加上
					itemContent = e.Domain + itemContent
				}
			} else {
				itemContent = s.Find(rule.Rule).Text()
			}
			item = append(item, KeyVal{Key: rule.Key, Val: itemContent})
		}
		parseRes = append(parseRes, item)
	})

	return parseRes, nil
}

// 解析json结果
func (e EngineParam) parseJson(httpBody []byte) ([][]KeyVal, error) {
	listData := gjson.GetBytes(httpBody, e.JsonRule.ListRule)
	if !listData.IsArray() {
		return nil, errors.New("json list rule res is not array, please check")
	}

	parseRes := make([][]KeyVal, 0, 10)
	listData.ForEach(func(key, value gjson.Result) bool {
		item := make([]KeyVal, 0, len(e.JsonRule.ListItemRule))
		for _, rule := range e.JsonRule.ListItemRule {
			itemContent := value.Get(rule.Rule).String()
			if rule.IsLink &&
				!strings.Contains(itemContent, "http://") &&
				!strings.Contains(itemContent, "https://") {
				// 如果链接不带域名, 需要加上
				itemContent = e.Domain + itemContent
			}
			item = append(item, KeyVal{Key: rule.Key, Val: itemContent})
		}
		parseRes = append(parseRes, item)
		return true
	})

	return parseRes, nil
}

func getEngineParamBing() EngineParam {
	return EngineParam{
		Desc:   "必应搜索",
		Domain: "https://www.bing.com",
		Param:  "/search?q={search_query}&ensearch=1",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".b_algo",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: "h2 a",
				},
				{
					Key:  "描述",
					Rule: "p",
				},
				{
					Key:  "链接",
					Rule: "h2 a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamBaidu() EngineParam {
	return EngineParam{
		Desc:   "百度搜索",
		Domain: "https://www.baidu.com",
		Param:  "/s?wd={search_query}",
		HtmlRule: &ParseHtmlRule{
			ListRule: "div.c-container.xpath-log.new-pmd",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".c-title a",
				},
				{
					Key:  "描述",
					Rule: ".content-right_8Zs40",
				},
				{
					Key:  "链接",
					Rule: ".c-title a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamGoogle() EngineParam {
	return EngineParam{
		Desc:   "谷歌搜索",
		Domain: "https://www.google.com",
		Param:  "/search?q={search_query}",
		//TODO google请求需要配置代理, 暂时不支持cli模式
		HtmlRule: &ParseHtmlRule{
			ListRule: ".MjjYud",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".yuRUbf a",
				},
				{
					Key:  "描述",
					Rule: "span.MUxGbd.wuQ4Ob.WZ8Tjf",
				},
				{
					Key:  "链接",
					Rule: ".yuRUbf a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamZhiHu() EngineParam {
	// 觉得bing带 site:zhihu.com 的搜索结果, 比知乎站内搜索好, 这里用bing搜索
	param := getEngineParamBing()
	param.Desc = "知乎搜索"
	param.Param = "/search?ensearch=0&q={search_query}" + url.QueryEscape(" site:zhihu.com")
	return param
}

func getEngineParamWeiXin() EngineParam {
	return EngineParam{
		Desc:   "搜狗微信搜索",
		Domain: "https://weixin.sogou.com",
		Param:  "/weixin?query={search_query}&type=2",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".txt-box",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: "h3 a",
				},
				{
					Key:  "描述",
					Rule: "p",
				},
				{
					Key:  "链接",
					Rule: "h3 a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamGithub() EngineParam {
	return EngineParam{
		Desc:   "Github搜索",
		Domain: "https://github.com",
		Param:  "/search?q={search_query}",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".repo-list-item",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".text-normal a",
				},
				{
					Key:  "描述",
					Rule: "p",
				},
				{
					Key:  "链接",
					Rule: ".text-normal a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamKaiFa() EngineParam {
	return EngineParam{
		Desc:    "百度开发者搜索",
		Domain:  "https://kaifa.baidu.com",
		Param:   "/searchPage?wd={search_query}",
		AjaxUrl: "https://kaifa.baidu.com/rest/v1/search?wd={search_query}&paramList=page_num%3D1%2Cpage_size%3D10&pageNum=1&pageSize=10",
		JsonRule: &ParseJsonRule{
			ListRule: "data.documents.data",
			ListItemRule: []ListItemJsonRule{
				{
					Key:  "标题",
					Rule: "techDocDigest.title",
				},
				{
					Key:  "描述",
					Rule: "techDocDigest.summary",
				},
				{
					Key:    "链接",
					Rule:   "techDocDigest.url",
					IsLink: true,
				},
			},
		},
	}
}

func getEngineParamDouBan() EngineParam {
	return EngineParam{
		Desc:   "豆瓣搜索",
		Domain: "https://www.douban.com",
		Param:  "/search?q={search_query}",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".result",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "名称",
					Rule: ".title a",
				},
				{
					Key:  "评分",
					Rule: ".rating-info .rating_nums",
				},
				{
					Key:  "评价人数",
					Rule: ".rating-info span:nth-child(3)",
				},
				{
					Key:  "出版信息",
					Rule: ".rating-info .subject-cast",
				},
				{
					Key:  "描述",
					Rule: "p",
				},
				{
					Key:  "链接",
					Rule: ".title a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamMovie() EngineParam {
	param := EngineParam{
		Desc:   "豆瓣电影搜索",
		Domain: "https://www.douban.com",
		Param:  "/search?cat=1002&q={search_query}",
	}
	douBanParam := getEngineParamDouBan()
	param.HtmlRule = douBanParam.HtmlRule

	return param
}

func getEngineParamBook() EngineParam {
	param := EngineParam{
		Desc:   "豆瓣书籍搜索",
		Domain: "https://www.douban.com",
		Param:  "/search?cat=1001&q={search_query}",
	}
	douBanParam := getEngineParamDouBan()
	param.HtmlRule = douBanParam.HtmlRule

	return param
}

func getEngineParam360() EngineParam {
	return EngineParam{
		Desc:   "360搜索",
		Domain: "https://www.so.com",
		Param:  "/s?q={search_query}",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".res-list",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".res-title a",
				},
				{
					Key:  "描述",
					Rule: ".res-desc",
				},
				{
					Key:  "链接",
					Rule: ".res-title a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamSoGou() EngineParam {
	return EngineParam{
		Desc:   "搜狗搜索",
		Domain: "https://www.sogou.com",
		Param:  "/web?query={search_query}",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".vrwrap",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".vr-title a",
				},
				{
					Key:  "描述",
					Rule: ".space-txt",
				},
				{
					Key:  "链接",
					Rule: ".vr-title a",
					Attr: "href",
				},
			},
		},
	}
}
