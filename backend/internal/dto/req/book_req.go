package req

import "time"

// CreateBookReq 创建书籍请求参数。
type CreateBookReq struct {
	Title         string     `json:"title" binding:"required,min=1,max=255"`
	Author        string     `json:"author" binding:"omitempty,max=255"`
	Cover         string     `json:"cover" binding:"omitempty,max=500"`
	Rating        *int       `json:"rating" binding:"omitempty,min=0,max=5"`
	ReadingStatus string     `json:"reading_status" binding:"omitempty,oneof=want reading done"`
	Review        string     `json:"review" binding:"omitempty,max=65535"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Status        *int       `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder     *int       `json:"sort_order"`
}

// UpdateBookReq 更新书籍请求参数（部分更新语义）。
type UpdateBookReq struct {
	Title         *string    `json:"title" binding:"omitempty,min=1,max=255"`
	Author        *string    `json:"author" binding:"omitempty,max=255"`
	Cover         *string    `json:"cover" binding:"omitempty,max=500"`
	Rating        *int       `json:"rating" binding:"omitempty,min=0,max=5"`
	ReadingStatus *string    `json:"reading_status" binding:"omitempty,oneof=want reading done"`
	Review        *string    `json:"review" binding:"omitempty,max=65535"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Status        *int       `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder     *int       `json:"sort_order"`
}

// BookListReq 书籍列表查询请求参数。
type BookListReq struct {
	PageReq
	Keyword       *string `form:"keyword" json:"keyword"`
	ReadingStatus *string `form:"reading_status" json:"reading_status"`
	Status        *int    `form:"status" json:"status"`
}
