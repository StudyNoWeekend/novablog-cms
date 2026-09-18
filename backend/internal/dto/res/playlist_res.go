package res

import "time"

// PlaylistRes 第三方歌单响应结构体。
type PlaylistRes struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	Platform    string    `json:"platform"`
	PlatformURL string    `json:"platform_url"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
