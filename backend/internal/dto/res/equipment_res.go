package res

import "time"

// EquipmentRes 个人设备响应结构体。
type EquipmentRes struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	// 图片媒体引用（CMS 编辑回显用）：媒体库文件 ID / 预设 ID，空=外链或历史 URL 模式
	ImageMediaID  string    `json:"image_media_id"`
	ImagePresetID string    `json:"image_preset_id"`
	Brand         string    `json:"brand"`
	Description   string    `json:"description"`
	SortOrder     int       `json:"sort_order"`
	ViewCount     int64     `json:"view_count"` // 浏览量
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// EquipmentListRes 个人设备列表响应结构体。
type EquipmentListRes struct {
	List       []EquipmentRes `json:"list"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
