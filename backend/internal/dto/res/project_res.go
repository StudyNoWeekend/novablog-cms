package res

import "time"

// ProjectRes 项目经历响应结构体。
type ProjectRes struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Role        string    `json:"role"`
	Client      string    `json:"client"`
	CoverURL    string    `json:"cover_url"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	TechStack   string    `json:"tech_stack"`
	StartDate   string    `json:"start_date"` // YYYY-MM，空表示未设置
	EndDate     string    `json:"end_date"`   // YYYY-MM，空表示至今或未设置
	ProjectURL  string    `json:"project_url"`
	RepoURL     string    `json:"repo_url"`
	Status      int       `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectListRes 项目经历列表响应结构体。
type ProjectListRes struct {
	List       []ProjectRes `json:"list"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}
