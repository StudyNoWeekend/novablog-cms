package res

import "time"

// SongRes 歌曲响应结构体。
type SongRes struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	CoverURL   string    `json:"cover_url"`
	BVID       string    `json:"bvid"`
	CID        int64     `json:"cid"`
	SourceURL  string    `json:"source_url"`
	SourceType string    `json:"source_type"`
	CategoryID *string   `json:"category_id"`
	Duration   int       `json:"duration"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ParseResultRes 解析结果响应结构体。
type ParseResultRes struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	CoverURL string `json:"cover_url"`
	Duration int    `json:"duration"`
	BVID     string `json:"bvid"`
	CID      int64  `json:"cid"`
}

// ParseTaskRes 解析任务响应结构体。
type ParseTaskRes struct {
	TaskID  string           `json:"task_id"`
	Status  string           `json:"status"` // pending, processing, success, failed
	Results []ParseResultRes `json:"results,omitempty"`
	Error   string           `json:"error,omitempty"`
}
