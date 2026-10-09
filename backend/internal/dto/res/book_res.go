package res

import "time"

// BookRes 书籍响应结构体（含书评，用于详情）。
type BookRes struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Author        string     `json:"author"`
	Cover         string     `json:"cover"`
	Rating        int        `json:"rating"`
	ReadingStatus string     `json:"reading_status"`
	Review        string     `json:"review"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Status        int        `json:"status"`
	SortOrder     int        `json:"sort_order"`
	ViewCount     int64      `json:"view_count"` // 浏览量
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BookCardRes 书籍列表响应结构体（不含书评大字段）。
type BookCardRes struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Cover         string    `json:"cover"`
	Rating        int       `json:"rating"`
	ReadingStatus string    `json:"reading_status"`
	Status        int       `json:"status"`
	SortOrder     int       `json:"sort_order"`
	ViewCount     int64     `json:"view_count"` // 浏览量
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BookListRes 书籍列表响应结构体。
type BookListRes struct {
	List       []BookCardRes `json:"list"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}
