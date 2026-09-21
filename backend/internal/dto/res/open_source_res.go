package res

import "time"

// OpenSourceRes 开源作品响应结构体（含 README，用于详情）。
type OpenSourceRes struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	RepoURL         string     `json:"repo_url"`
	Summary         string     `json:"summary"`
	Readme          string     `json:"readme"` // README 原文（Markdown）
	Language        string     `json:"language"`
	Topics          string     `json:"topics"`
	Stars           int        `json:"stars"`
	Homepage        string     `json:"homepage"`
	Status          int        `json:"status"`
	SortOrder       int        `json:"sort_order"`
	ReadmeUpdatedAt *time.Time `json:"readme_updated_at"` // README 最近拉取时间，null 表示尚未拉取
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// OpenSourceCardRes 开源作品列表响应结构体（不含 README 大字段）。
type OpenSourceCardRes struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	Summary   string    `json:"summary"`
	Language  string    `json:"language"`
	Topics    string    `json:"topics"`
	Stars     int       `json:"stars"`
	Homepage  string    `json:"homepage"`
	Status    int       `json:"status"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OpenSourceListRes 开源作品列表响应结构体。
type OpenSourceListRes struct {
	List       []OpenSourceCardRes `json:"list"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}
