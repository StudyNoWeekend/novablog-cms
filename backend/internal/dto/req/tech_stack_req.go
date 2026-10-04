package req

// CreateTechStackReq 创建技术栈条目请求参数。
type CreateTechStackReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Category    string `json:"category" binding:"omitempty,max=50"` // language/framework/tool/database 等
	Icon        string `json:"icon" binding:"omitempty,max=500"`
	Level       *int   `json:"level" binding:"omitempty,min=1,max=4"` // 1=了解, 2=熟悉, 3=熟练, 4=精通
	Description string `json:"description" binding:"omitempty,max=500"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int   `json:"sort_order"`
}

// UpdateTechStackReq 更新技术栈条目请求参数（部分更新语义）。
type UpdateTechStackReq struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Category    *string `json:"category" binding:"omitempty,max=50"`
	Icon        *string `json:"icon" binding:"omitempty,max=500"`
	Level       *int    `json:"level" binding:"omitempty,min=1,max=4"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int    `json:"sort_order"`
}

// TechStackListReq 技术栈条目列表查询请求参数。
type TechStackListReq struct {
	PageReq
	Keyword  *string `form:"keyword" json:"keyword"`
	Category *string `form:"category" json:"category"`
	Level    *int    `form:"level" json:"level"`
	Status   *int    `form:"status" json:"status"`
}
