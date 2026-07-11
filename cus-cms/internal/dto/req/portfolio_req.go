package req

// CreatePortfolioReq 创建作品集请求参数。
type CreatePortfolioReq struct {
	Name          string  `json:"name" binding:"required,min=1,max=255"`
	Description   string  `json:"description"`
	CoverMode     *int    `json:"cover_mode"`      // 0=使用排序第一的作品, 1=独立设置封面
	CoverPresetID *string `json:"cover_preset_id"` // cover_mode=1 时必填
	Status        *int    `json:"status"`
	SortOrder     *int    `json:"sort_order"`
	CategoryID    *string `json:"category_id"`
}

// UpdatePortfolioReq 更新作品集请求参数。
type UpdatePortfolioReq struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	CoverMode     *int    `json:"cover_mode"`      // 0=使用排序第一的作品, 1=独立设置封面
	CoverPresetID *string `json:"cover_preset_id"` // cover_mode=1 时必填
	Status        *int    `json:"status"`
	SortOrder     *int    `json:"sort_order"`
	CategoryID    *string `json:"category_id"`
}

// PortfolioListReq 作品集列表查询请求参数。
type PortfolioListReq struct {
	PageReq
	Keyword    *string `form:"keyword" json:"keyword"`
	Status     *int    `form:"status" json:"status"`
	CategoryID *string `form:"category_id" json:"category_id"`
}

// CreatePortfolioItemReq 添加作品项请求参数。
type CreatePortfolioItemReq struct {
	PresetID    string `json:"preset_id" binding:"required"`
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description"`
}

// UpdatePortfolioItemReq 更新作品项请求参数。
type UpdatePortfolioItemReq struct {
	PresetID    *string `json:"preset_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// SortPortfolioItemReq 排序项请求参数。
type SortPortfolioItemReq struct {
	ID        string `json:"id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

// SortPortfolioItemsReq 批量排序作品项请求参数。
type SortPortfolioItemsReq struct {
	Items []SortPortfolioItemReq `json:"items" binding:"required,min=1"`
}
