package req

import "time"

// CreateFitnessReq 创建训练记录请求参数。
type CreateFitnessReq struct {
	Date        *time.Time       `json:"date"`
	Title       string           `json:"title" binding:"required,min=1,max=255"`
	CategoryID  *string          `json:"category_id"`                                            // 结构化分类（type=fitness），设置后 Type 由后端镜像为分类 slug
	Type        string           `json:"type" binding:"omitempty,oneof=strength cardio stretch"` // 兼容保留：建议改用 category_id
	DurationMin *int             `json:"duration_min" binding:"omitempty,min=0"`
	Calories    *int             `json:"calories" binding:"omitempty,min=0"`
	Content     []map[string]any `json:"content"` // 动作清单 [{name,sets,reps,note}]
	Notes       string           `json:"notes" binding:"omitempty,max=65535"`
	Status      *int             `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int             `json:"sort_order"`
}

// UpdateFitnessReq 更新训练记录请求参数（部分更新语义）。
type UpdateFitnessReq struct {
	Date        *time.Time       `json:"date"`
	Title       *string          `json:"title" binding:"omitempty,min=1,max=255"`
	CategoryID  *string          `json:"category_id"`
	Type        *string          `json:"type" binding:"omitempty,oneof=strength cardio stretch"` // 兼容保留，建议用 category_id
	DurationMin *int             `json:"duration_min" binding:"omitempty,min=0"`
	Calories    *int             `json:"calories" binding:"omitempty,min=0"`
	Content     []map[string]any `json:"content"`
	Notes       *string          `json:"notes" binding:"omitempty,max=65535"`
	Status      *int             `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int             `json:"sort_order"`
}

// FitnessListReq 训练记录列表查询请求参数。
type FitnessListReq struct {
	PageReq
	Keyword    *string `form:"keyword" json:"keyword"`
	CategoryID *string `form:"category_id" json:"category_id"`
	Type       *string `form:"type" json:"type"` // 兼容保留，按 legacy 镜像值筛选
	Status     *int    `form:"status" json:"status"`
}
