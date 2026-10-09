package res

import "time"

// RecipeRes 菜谱响应结构体（含食材/步骤，用于详情）。
type RecipeRes struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Cover       string           `json:"cover"`
	Summary     string           `json:"summary"`
	Ingredients []map[string]any `json:"ingredients"`
	Steps       []map[string]any `json:"steps"`
	Difficulty  int              `json:"difficulty"`
	Minutes     int              `json:"minutes"`
	Servings    int              `json:"servings"`
	Tags        string           `json:"tags"`
	Status      int              `json:"status"`
	SortOrder   int              `json:"sort_order"`
	ViewCount   int64            `json:"view_count"` // 浏览量
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// RecipeCardRes 菜谱列表响应结构体（不含食材/步骤大字段）。
type RecipeCardRes struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Cover      string    `json:"cover"`
	Summary    string    `json:"summary"`
	Difficulty int       `json:"difficulty"`
	Minutes    int       `json:"minutes"`
	Servings   int       `json:"servings"`
	Tags       string    `json:"tags"`
	Status     int       `json:"status"`
	SortOrder  int       `json:"sort_order"`
	ViewCount  int64     `json:"view_count"` // 浏览量
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RecipeListRes 菜谱列表响应结构体。
type RecipeListRes struct {
	List       []RecipeCardRes `json:"list"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}
