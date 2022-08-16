package search

import (
	"fmt"
	"net/url"
	"strings"
)

// 搜索引擎
const (
	EngineBing   = "bing"   // bing搜索
	EngineBaidu  = "baidu"  // baidu搜索
	EngineGoogle = "google" // google搜索
	EngineZhiHu  = "zhihu"  // zhihu搜索
	EngineWeiXin = "weixin" // 微信搜索
	EngineGithub = "github" // github搜索
	EngineKaiFa  = "kaifa"  // baidu开发者搜索
	EngineDouBan = "douban" // 豆瓣搜索
	EngineMovie  = "movie"  // 豆瓣电影搜索
	EngineBook   = "book"   // 豆瓣书籍搜索
	Engine360    = "360"    // 360搜索
	EngineSoGou  = "sogou"  // 搜狗搜索
)

// EngineParam 搜索引擎参数
type EngineParam struct {
	Desc  string // 说明
	Url   string // 网址
	Param string // 检索参数
}

// EngineParamMap 搜索引擎映射
var EngineParamMap = map[string]EngineParam{
	EngineBing: {
		Desc:  "必应搜索",
		Url:   "https://cn.bing.com",
		Param: "/search?q=%s&ensearch=1",
	},
	EngineBaidu: {
		Desc:  "百度搜索",
		Url:   "https://www.baidu.com",
		Param: "/s?wd=%s",
	},
	EngineGoogle: {
		Desc:  "谷歌搜索",
		Url:   "https://www.google.com",
		Param: "/search?q=%s",
	},
	EngineZhiHu: {
		Desc:  "知乎搜索",
		Url:   "https://www.zhihu.com",
		Param: "/search?q=%s&type=content",
	},
	EngineWeiXin: {
		Desc:  "搜狗微信搜索",
		Url:   "https://weixin.sogou.com",
		Param: "/weixin?query=%s&type=2",
	},
	EngineGithub: {
		Desc:  "Github搜索",
		Url:   "https://github.com",
		Param: "/search?q=%s",
	},
	EngineKaiFa: {
		Desc:  "百度开发者搜索",
		Url:   "https://kaifa.baidu.com",
		Param: "/searchPage?wd=%s",
	},
	EngineDouBan: {
		Desc:  "豆瓣搜索",
		Url:   "https://www.douban.com",
		Param: "/search?q=%s",
	},
	EngineMovie: {
		Desc:  "豆瓣电影搜索",
		Url:   "https://www.douban.com",
		Param: "/search?cat=1002&q=%s",
	},
	EngineBook: {
		Desc:  "豆瓣书籍搜索",
		Url:   "https://www.douban.com",
		Param: "/search?cat=1001&q=%s",
	},
	Engine360: {
		Desc:  "360搜索",
		Url:   "https://www.so.com",
		Param: "/s?q=%s",
	},
	EngineSoGou: {
		Desc:  "搜狗搜索",
		Url:   "https://www.sogou.com",
		Param: "/web?query=%s",
	},
}

// FormatSearchUrl 格式化检索网址
func FormatSearchUrl(searchEngine string, query string) string {
	engineParam, ok := EngineParamMap[searchEngine]

	// 如果没有对应的搜索引擎, 使用必应
	if !ok {
		engineParam = EngineParamMap[EngineBing]
	}

	// 如果没有设置检索query, 只打开搜索引擎
	if len(query) <= 0 {
		return engineParam.Url
	}

	return engineParam.Url + fmt.Sprintf(engineParam.Param, url.QueryEscape(query))
}

func FormatCommandDesc() string {
	commandDesc := make([]string, 0, len(EngineParamMap)+1)
	commandDesc = append(commandDesc, "打开默认浏览器, 指定搜索引擎, 检索相关query，模式如下：")
	for engineName, engineParam := range EngineParamMap {
		commandDesc = append(commandDesc, engineName+": "+engineParam.Desc)
	}
	return strings.Join(commandDesc, "\n")
}
