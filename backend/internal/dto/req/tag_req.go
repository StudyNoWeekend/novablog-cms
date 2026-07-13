package req

// CreateTagReq 创建标签请求参数。
type CreateTagReq struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateTagReq 更新标签请求参数。
type UpdateTagReq struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}
