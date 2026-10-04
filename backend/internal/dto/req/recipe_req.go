package req

// CreateRecipeReq 创建菜谱请求参数。
type CreateRecipeReq struct {
	Title       string           `json:"title" binding:"required,min=1,max=255"`
	Cover       string           `json:"cover" binding:"omitempty,max=500"`
	Summary     string           `json:"summary" binding:"omitempty,max=500"`
	Ingredients []map[string]any `json:"ingredients"`                                // 食材清单 [{name,amount}]
	Steps       []map[string]any `json:"steps"`                                      // 步骤 [{text,image}]
	Difficulty  *int             `json:"difficulty" binding:"omitempty,min=1,max=3"` // 1=简单, 2=中等, 3=困难
	Minutes     *int             `json:"minutes" binding:"omitempty,min=0"`
	Servings    *int             `json:"servings" binding:"omitempty,min=1"`
	Tags        string           `json:"tags" binding:"omitempty,max=500"` // 标签，逗号分隔
	Status      *int             `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int             `json:"sort_order"`
}

// UpdateRecipeReq 更新菜谱请求参数（部分更新语义）。
type UpdateRecipeReq struct {
	Title       *string          `json:"title" binding:"omitempty,min=1,max=255"`
	Cover       *string          `json:"cover" binding:"omitempty,max=500"`
	Summary     *string          `json:"summary" binding:"omitempty,max=500"`
	Ingredients []map[string]any `json:"ingredients"`
	Steps       []map[string]any `json:"steps"`
	Difficulty  *int             `json:"difficulty" binding:"omitempty,min=1,max=3"`
	Minutes     *int             `json:"minutes" binding:"omitempty,min=0"`
	Servings    *int             `json:"servings" binding:"omitempty,min=1"`
	Tags        *string          `json:"tags" binding:"omitempty,max=500"`
	Status      *int             `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int             `json:"sort_order"`
}

// RecipeListReq 菜谱列表查询请求参数。
type RecipeListReq struct {
	PageReq
	Keyword    *string `form:"keyword" json:"keyword"`
	Difficulty *int    `form:"difficulty" json:"difficulty"`
	Status     *int    `form:"status" json:"status"`
}
