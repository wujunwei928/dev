package search

import (
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
	Desc     string // 说明
	Domain   string // 浏览器检索域名
	Param    string // 浏览器检索参数
	AjaxUrl  string // ajax请求地址: 部分网站是使用ajax渲染, 终端模式下需请求此地址
	Cookie   string // cookie
	HtmlRule *ParseHtmlRule
	JsonRule *ParseJsonRule
}

// EngineParamMap 搜索引擎映射
var EngineParamMap = map[string]EngineParam{
	EngineBing:   getEngineParamBing(),
	EngineBaidu:  getEngineParamBaidu(),
	EngineGoogle: getEngineParamGoogle(),
	EngineZhiHu:  getEngineParamZhiHu(),
	EngineWeiXin: getEngineParamWeiXin(),
	EngineGithub: getEngineParamGithub(),
	EngineKaiFa:  getEngineParamKaiFa(),
	EngineDouBan: getEngineParamDouBan(),
	EngineMovie:  getEngineParamMovie(),
	EngineBook:   getEngineParamBook(),
	Engine360:    getEngineParam360(),
	EngineSoGou:  getEngineParamSoGou(),
}

// 获取搜索引擎参数
func getEngineParam(searchEngine string) EngineParam {
	engineParam, ok := EngineParamMap[searchEngine]

	// 如果没有对应的搜索引擎, 使用必应
	if !ok {
		engineParam = EngineParamMap[EngineBing]
	}

	return engineParam
}

// FormatSearchUrl 格式化检索网址
func FormatSearchUrl(searchEngine string, query string) string {
	engineParam := getEngineParam(searchEngine)

	// 如果没有设置检索query, 只打开搜索引擎
	if len(query) <= 0 {
		return engineParam.Domain
	}

	return engineParam.Domain + strings.ReplaceAll(engineParam.Param, "{search_query}", url.QueryEscape(query))
}

func FormatSearchCommandModeUsage() string {
	commandDesc := make([]string, 0, len(EngineParamMap)+1)
	commandDesc = append(commandDesc, "指定搜索引擎, 支持搜索引擎如下：")
	for engineName, engineParam := range EngineParamMap {
		commandDesc = append(commandDesc, engineName+": "+engineParam.Desc)
	}
	return strings.Join(commandDesc, "\n")
}
