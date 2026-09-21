package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// PhotoEquipment 个人设备模型，对应 photo_equipment 数据表。
type PhotoEquipment struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"type:varchar(255);not null"`
	ImageURL    string         `gorm:"column:image_url;type:varchar(1024)"`
	Brand       string         `gorm:"type:varchar(255)"`
	Description string         `gorm:"type:text"`
	SortOrder   int            `gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (PhotoEquipment) TableName() string {
	return "photo_equipment"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (e *PhotoEquipment) AfterFind(tx *gorm.DB) error {
	e.ImageURL = resolveURL(e.ImageURL)
	return nil
}

// PhotoEquipmentModel 个人设备模型操作结构体。
type PhotoEquipmentModel struct {
	db *gorm.DB
}

// NewPhotoEquipment 创建 PhotoEquipmentModel 实例。
func NewPhotoEquipment() *PhotoEquipmentModel {
	return &PhotoEquipmentModel{db: DB}
}

// Create 创建个人设备记录。
func (m *PhotoEquipmentModel) Create(ctx context.Context, e *PhotoEquipment) error {
	return m.db.WithContext(ctx).Create(e).Error
}

// GetByID 根据 ID 查询个人设备。
func (m *PhotoEquipmentModel) GetByID(ctx context.Context, id string) (*PhotoEquipment, error) {
	var e PhotoEquipment
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetList 分页查询个人设备列表，支持按 keyword 模糊搜索名称、按 brand 过滤。
func (m *PhotoEquipmentModel) GetList(ctx context.Context, keyword *string, brand *string, page, pageSize int) ([]PhotoEquipment, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&PhotoEquipment{})

	if keyword != nil && *keyword != "" {
		query = query.Where("name ILIKE ?", "%"+*keyword+"%")
	}
	if brand != nil && *brand != "" {
		query = query.Where("brand ILIKE ?", "%"+*brand+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []PhotoEquipment
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新个人设备。
func (m *PhotoEquipmentModel) Update(ctx context.Context, e *PhotoEquipment) error {
	return m.db.WithContext(ctx).Save(e).Error
}

// SoftDelete 软删除个人设备。
func (m *PhotoEquipmentModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&PhotoEquipment{}).Error
}
