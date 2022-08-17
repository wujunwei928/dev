package search

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// KeyVal Golang中map遍历没有顺序, 使用[]KeyVal代替map[string]string返回
type KeyVal struct {
	Key string
	Val string
}

func RequestDetail(searchEngine string, query string) ([][]KeyVal, error) {
	engineParam := getEngineParam(searchEngine)

	reqUrl := engineParam.Domain + fmt.Sprintf(engineParam.Param, url.QueryEscape(query))
	referUrl := engineParam.Domain
	if searchEngine == EngineZhiHu {
		referUrl = reqUrl
		reqUrl = "https://www.zhihu.com/api/v4/search_v3?gk_version=gz-gaokao&t=general&q=php&correction=1&offset=0&limit=20&filter_fields=&lc_idx=0&show_all_topics=0&search_source=Normal"
	}
	client := resty.New()
	client.SetTimeout(2000 * time.Millisecond)
	res, err := client.R().
		EnableTrace().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.81 Safari/537.36 Edg/104.0.1293.54").
		SetHeader("Referer", referUrl).
		SetHeader("Cookie", "").
		Get(reqUrl)
	if err != nil {
		fmt.Println(pterm.Red(err.Error()))
		return nil, err
	}
	fmt.Println(res.Request, string(res.Body()))

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return nil, err
	}

	parseRes := make([][]KeyVal, 0, 10)
	//fmt.Println(engineParam.HtmlRule.ListRule)
	doc.Find(engineParam.HtmlRule.ListRule).Each(func(i int, s *goquery.Selection) {
		item := make([]KeyVal, 0, len(engineParam.HtmlRule.ListItemRule))
		for _, rule := range engineParam.HtmlRule.ListItemRule {
			var itemContent string
			if len(rule.Attr) > 0 {
				itemContent, _ = s.Find(rule.Rule).Attr(rule.Attr)
				if rule.Attr == "href" &&
					!strings.Contains(itemContent, "http://") &&
					!strings.Contains(itemContent, "https://") {
					// 有些网站(如: 搜狗)的跳转链接不带域名, 需要加上
					itemContent = engineParam.Domain + itemContent
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

func getEngineParamBing() EngineParam {
	return EngineParam{
		Desc:   "必应搜索",
		Domain: "https://cn.bing.com",
		Param:  "/search?q=%s&ensearch=1",
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
		Param:  "/s?wd=%s",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".c-container",
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
		Param:  "/search?q=%s",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".c-container",
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

func getEngineParamZhiHu() EngineParam {
	return EngineParam{
		Desc:    "知乎搜索",
		Domain:  "https://www.zhihu.com",
		Param:   "/search?q=%s&type=content",
		AjaxUrl: "https://kaifa.baidu.com/rest/v1/search?wd=php%20substr&paramList=page_num%3D1%2Cpage_size%3D10&pageNum=1&pageSize=10",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".SearchResult-Card",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: "h2.ContentItem-title > div > a",
				},
				{
					Key:  "描述",
					Rule: ".CopyrightRichText-richText",
				},
				{
					Key:  "链接",
					Rule: "h2.ContentItem-title > div > a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamWeiXin() EngineParam {
	return EngineParam{
		Desc:   "搜狗微信搜索",
		Domain: "https://weixin.sogou.com",
		Param:  "/weixin?query=%s&type=2",
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
		Param:  "/search?q=%s",
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
		Param:   "/searchPage?wd=%s",
		AjaxUrl: "",
		HtmlRule: &ParseHtmlRule{
			ListRule: ".ant-list-item",
			ListItemRule: []ListItemHtmlRule{
				{
					Key:  "标题",
					Rule: ".title-root-e6482 a",
				},
				{
					Key:  "描述",
					Rule: ".summary-root-b31c4 span",
				},
				{
					Key:  "链接",
					Rule: ".title-root-e6482 a",
					Attr: "href",
				},
			},
		},
	}
}

func getEngineParamDouBan() EngineParam {
	return EngineParam{
		Desc:   "豆瓣搜索",
		Domain: "https://www.douban.com",
		Param:  "/search?q=%s",
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
		Param:  "/search?cat=1002&q=%s",
	}
	douBanParam := getEngineParamDouBan()
	param.HtmlRule = douBanParam.HtmlRule

	return param
}

func getEngineParamBook() EngineParam {
	param := EngineParam{
		Desc:   "豆瓣书籍搜索",
		Domain: "https://www.douban.com",
		Param:  "/search?cat=1001&q=%s",
	}
	douBanParam := getEngineParamDouBan()
	param.HtmlRule = douBanParam.HtmlRule

	return param
}

func getEngineParam360() EngineParam {
	return EngineParam{
		Desc:   "360搜索",
		Domain: "https://www.so.com",
		Param:  "/s?q=%s",
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
		Param:  "/web?query=%s",
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
