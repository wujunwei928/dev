package search

// ParseJsonRule json解析规则
type ParseJsonRule struct {
	ListRule     string             // 列表解析规则, 循环解析获取单挑属性值
	ListItemRule []ListItemJsonRule // 列表单条元素解析规则, 在ListRule的基础上向下解析
}

// ListItemJsonRule json列表元素解析规则
type ListItemJsonRule struct {
	Key    string // 解析后的key
	Rule   string // 解析规则
	IsLink bool   // 是否链接: 判断是否需要加域名(个别网站返回的链接可能没有域名, 需要补全)
}
