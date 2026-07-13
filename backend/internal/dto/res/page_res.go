package res

// PageRes 通用分页响应结构体。
type PageRes[T any] struct {
	List       []T   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// NewPageRes 创建分页响应。
func NewPageRes[T any](list []T, total int64, page, pageSize int) *PageRes[T] {
	if list == nil {
		list = []T{}
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &PageRes[T]{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
