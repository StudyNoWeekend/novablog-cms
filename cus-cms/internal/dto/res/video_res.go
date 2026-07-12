package res

import "time"

// VideoPlatformLinkRes 视频平台链接响应结构体。
type VideoPlatformLinkRes struct {
	ID        string    `json:"id"`
	VideoID   string    `json:"video_id"`
	Platform  string    `json:"platform"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VideoWorkRes 视频作品响应结构体。
type VideoWorkRes struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	CoverURL    string                 `json:"cover_url"`
	Description string                 `json:"description"`
	Status      int                    `json:"status"`
	SortOrder   int                    `json:"sort_order"`
	Platforms   []VideoPlatformLinkRes `json:"platforms"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// VideoWorkListRes 视频作品列表响应结构体。
type VideoWorkListRes struct {
	List       []VideoWorkRes `json:"list"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ParseVideoRes 视频元信息解析响应结构体。
type ParseVideoRes struct {
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
	Description string `json:"description"`
}
