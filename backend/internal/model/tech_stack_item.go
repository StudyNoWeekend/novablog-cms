package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TechStackItem 技术栈条目模型，对应 tech_stack_items 数据表。
type TechStackItem struct {
	ID           string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Category     string         `gorm:"type:varchar(50);default:''" json:"category"` // 兼容镜像：通常等于关联分类的 slug（language/framework/tool/database 等）
	Icon         string         `gorm:"type:varchar(500)" json:"icon"`               // 图标 URL
	// 图标媒体引用：存媒体库文件 ID 时 icon 留空，读取时动态解析；Simple Icons 等外链仍存 icon
	IconMediaID *string `gorm:"column:icon_media_id;type:uuid;index" json:"icon_media_id"`
	Level        int            `gorm:"type:int;default:1" json:"level"`             // 1=了解, 2=熟悉, 3=熟练, 4=精通
	Description  string         `gorm:"type:varchar(500)" json:"description"`
	Status       int            `gorm:"type:int;default:0" json:"status"` // 0=草稿, 1=已发布
	SortOrder    int            `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	ViewCount    int            `gorm:"column:view_count;type:int;default:0" json:"view_count"`
	CategoryID   *string        `gorm:"type:uuid;column:category_id" json:"category_id"`
	CategoryInfo *Category      `gorm:"foreignKey:CategoryID" json:"-"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave 保存前归一图标引用：ID 模式下清空 URL 列，避免解析出的完整 URL 被写回落库。
func (t *TechStackItem) BeforeSave(tx *gorm.DB) error {
	if t.IconMediaID != nil {
		t.Icon = ""
	}
	return nil
}

// TableName 指定数据表名称。
func (TechStackItem) TableName() string {
	return "tech_stack_items"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (t *TechStackItem) AfterFind(tx *gorm.DB) error {
	t.Icon = resolveURL(t.Icon)
	return nil
}

// TechStackItemModel 技术栈条目模型操作结构体。
type TechStackItemModel struct {
	db *gorm.DB
}

// NewTechStackItem 创建 TechStackItemModel 实例。
func NewTechStackItem() *TechStackItemModel {
	return &TechStackItemModel{db: DB}
}

// Create 创建技术栈条目。
func (m *TechStackItemModel) Create(ctx context.Context, t *TechStackItem) error {
	return m.db.WithContext(ctx).Create(t).Error
}

// GetByID 根据 ID 查询技术栈条目。
func (m *TechStackItemModel) GetByID(ctx context.Context, id string) (*TechStackItem, error) {
	var t TechStackItem
	err := m.db.WithContext(ctx).Preload("CategoryInfo").Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetList 分页查询技术栈条目列表，支持 keyword 模糊搜索名称、category 筛选（兼容镜像）、
// 按分类过滤、level 筛选、status 过滤。
func (m *TechStackItemModel) GetList(ctx context.Context, keyword *string, category *string, categoryID *string, level *int, status *int, page, pageSize int) ([]TechStackItem, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&TechStackItem{})

	if keyword != nil && *keyword != "" {
		query = query.Where("name ILIKE ?", "%"+*keyword+"%")
	}
	if category != nil && *category != "" {
		query = query.Where("category = ?", *category)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}
	if level != nil && *level > 0 {
		query = query.Where("level = ?", *level)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []TechStackItem
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Preload("CategoryInfo").
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新技术栈条目。
func (m *TechStackItemModel) Update(ctx context.Context, t *TechStackItem) error {
	return m.db.WithContext(ctx).Save(t).Error
}

// SoftDelete 软删除技术栈条目。
func (m *TechStackItemModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&TechStackItem{}).Error
}
