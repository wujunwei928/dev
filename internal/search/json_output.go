package search

import "encoding/json"

// SearchResult JSON 模式的搜索结果
type SearchResult struct {
	Engine  string       `json:"engine"`
	Query   string       `json:"query"`
	Results []ResultItem `json:"results"`
}

// ResultItem 单条搜索结果
type ResultItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// keyToFieldMapping 中文 key 到英文字段名的映射
var keyToFieldMapping = map[string]string{
	"标题":   "title",
	"名称":   "title",
	"链接":   "url",
	"描述":   "description",
	"评分":   "",
	"评价人数": "",
	"出版信息": "",
}

// KeyValToResultItems 将 [][]KeyVal 转换为 SearchResult
func KeyValToResultItems(engine string, query string, kvResults [][]KeyVal) SearchResult {
	items := make([]ResultItem, 0, len(kvResults))
	for _, kvRow := range kvResults {
		item := ResultItem{}
		for _, kv := range kvRow {
			field := keyToFieldMapping[kv.Key]
			switch field {
			case "title":
				item.Title = kv.Val
			case "url":
				item.URL = kv.Val
			case "description":
				item.Description = kv.Val
			}
		}
		items = append(items, item)
	}
	return SearchResult{
		Engine:  engine,
		Query:   query,
		Results: items,
	}
}

// ToJSON 将搜索结果序列化为 JSON 字符串
func (sr SearchResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(sr, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
