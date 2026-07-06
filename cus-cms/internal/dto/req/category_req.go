package req

// CreateCategoryReq 创建分类请求参数。
type CreateCategoryReq struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateCategoryReq 更新分类请求参数。
type UpdateCategoryReq struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}