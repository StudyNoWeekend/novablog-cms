package req

// CreateEquipmentReq 创建个人设备请求参数。
type CreateEquipmentReq struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	ImageURL    string `json:"image_url"`
	Brand       string `json:"brand"`
	Description string `json:"description"`
}

// UpdateEquipmentReq 更新个人设备请求参数。
type UpdateEquipmentReq struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	ImageURL    *string `json:"image_url"`
	Brand       *string `json:"brand"`
	Description *string `json:"description"`
}

// EquipmentListReq 个人设备列表查询请求参数。
type EquipmentListReq struct {
	PageReq
	Keyword *string `form:"keyword" json:"keyword"`
	Brand   *string `form:"brand" json:"brand"`
}
