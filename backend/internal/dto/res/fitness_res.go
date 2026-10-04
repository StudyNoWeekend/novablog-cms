package res

import "time"

// FitnessRes 训练记录响应结构体（含动作清单，用于详情）。
type FitnessRes struct {
	ID          string           `json:"id"`
	Date        time.Time        `json:"date"`
	Title       string           `json:"title"`
	Type        string           `json:"type"`
	DurationMin int              `json:"duration_min"`
	Calories    int              `json:"calories"`
	Content     []map[string]any `json:"content"`
	Notes       string           `json:"notes"`
	Status      int              `json:"status"`
	SortOrder   int              `json:"sort_order"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// FitnessCardRes 训练记录列表响应结构体（不含动作清单大字段）。
type FitnessCardRes struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	DurationMin int       `json:"duration_min"`
	Calories    int       `json:"calories"`
	Status      int       `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FitnessListRes 训练记录列表响应结构体。
type FitnessListRes struct {
	List       []FitnessCardRes `json:"list"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}
