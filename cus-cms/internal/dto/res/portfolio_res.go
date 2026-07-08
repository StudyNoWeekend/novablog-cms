package res

import "time"

// PortfolioRes 作品集响应结构体。
type PortfolioRes struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CoverMode     int       `json:"cover_mode"`      // 0=使用排序第一的作品, 1=独立设置封面
	CoverPresetID string    `json:"cover_preset_id"` // 独立封面预设ID
	CoverURL      string    `json:"cover_url"`       // 解析后的封面图地址
	Status        int       `json:"status"`
	SortOrder     int       `json:"sort_order"`
	ItemCount     int64     `json:"item_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PortfolioListRes 作品集列表响应结构体。
type PortfolioListRes struct {
	List       []PortfolioRes `json:"list"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// PortfolioItemRes 作品项响应结构体。
type PortfolioItemRes struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolio_id"`
	PresetID    string    `json:"preset_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	OutputURL   string    `json:"output_url"`
	MimeType    string    `json:"mime_type"`
	OutputSize  int64     `json:"output_size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PortfolioDetailRes 作品集详情响应结构体。
type PortfolioDetailRes struct {
	PortfolioRes
	Items []PortfolioItemRes `json:"items"`
}
