package search

// ParseHtmlRule html解析规则
type ParseHtmlRule struct {
	ListRule     string             // 列表解析规则, 循环解析获取单挑属性值
	ListItemRule []ListItemHtmlRule // 列表单条元素解析规则, 在ListRule的基础上向下解析
}

// ListItemHtmlRule html列表元素解析规则
type ListItemHtmlRule struct {
	Key  string // 解析后的key
	Rule string // 解析规则
	Attr string // 属性: 不设置时, 取text值; 设置时取对应的属性值
}
