package res

import "time"

// TechStackItemRes 技术栈条目响应结构体。
type TechStackItemRes struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Icon        string    `json:"icon"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TechStackListRes 技术栈条目列表响应结构体。
type TechStackListRes struct {
	List       []TechStackItemRes `json:"list"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
