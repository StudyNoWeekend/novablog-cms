package req

// CreatePortfolioReq 创建作品集请求参数。
type CreatePortfolioReq struct {
	Name          string  `json:"name" binding:"required,min=1,max=255"`
	Description   string  `json:"description"`
	CoverPresetID *string `json:"cover_preset_id"`
	Status        *int    `json:"status"`
	SortOrder     *int    `json:"sort_order"`
}

// UpdatePortfolioReq 更新作品集请求参数。
type UpdatePortfolioReq struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	CoverPresetID *string `json:"cover_preset_id"`
	Status        *int    `json:"status"`
	SortOrder     *int    `json:"sort_order"`
}

// PortfolioListReq 作品集列表查询请求参数。
type PortfolioListReq struct {
	PageReq
	Keyword *string `form:"keyword" json:"keyword"`
	Status  *int    `form:"status" json:"status"`
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
