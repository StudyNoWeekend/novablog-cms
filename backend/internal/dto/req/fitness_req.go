package req

import "time"

// CreateFitnessReq 创建训练记录请求参数。
type CreateFitnessReq struct {
	Date        *time.Time       `json:"date"`
	Title       string           `json:"title" binding:"required,min=1,max=255"`
	Type        string           `json:"type" binding:"omitempty,oneof=strength cardio stretch"`
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
	Type        *string          `json:"type" binding:"omitempty,oneof=strength cardio stretch"`
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
	Keyword *string `form:"keyword" json:"keyword"`
	Type    *string `form:"type" json:"type"`
	Status  *int    `form:"status" json:"status"`
}
