package types

type Pageable struct {
	Records interface{} `json:"records"` // 数据集合，支持任意类型
	Total   int64       `json:"total"`   // 总记录数
	Page    int         `json:"page"`    // 当前页
	Limit   int         `json:"limit"`   // 每页数据
}
