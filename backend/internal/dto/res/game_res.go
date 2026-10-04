package res

import "time"

// GameRes 游戏响应结构体（含短评，用于详情）。
type GameRes struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Cover       string    `json:"cover"`
	Platform    string    `json:"platform"`
	Genre       string    `json:"genre"`
	PlayStatus  string    `json:"play_status"`
	PlayHours   int       `json:"play_hours"`
	Rating      int       `json:"rating"`
	ShortReview string    `json:"short_review"`
	Status      int       `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GameCardRes 游戏列表响应结构体（不含短评大字段）。
type GameCardRes struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Cover      string    `json:"cover"`
	Platform   string    `json:"platform"`
	Genre      string    `json:"genre"`
	PlayStatus string    `json:"play_status"`
	PlayHours  int       `json:"play_hours"`
	Rating     int       `json:"rating"`
	Status     int       `json:"status"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GameListRes 游戏列表响应结构体。
type GameListRes struct {
	List       []GameCardRes `json:"list"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}
