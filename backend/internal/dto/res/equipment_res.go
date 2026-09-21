package res

import "time"

// EquipmentRes 个人设备响应结构体。
type EquipmentRes struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	Brand       string    `json:"brand"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EquipmentListRes 个人设备列表响应结构体。
type EquipmentListRes struct {
	List       []EquipmentRes `json:"list"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
