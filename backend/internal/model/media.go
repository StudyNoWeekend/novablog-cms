package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Media 媒体文件模型，对应 media 数据表。
type Media struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Filename    string         `gorm:"type:varchar(255);not null"`
	FileType    int16          `gorm:"column:file_type;not null"`
	MimeType    string         `gorm:"column:mime_type;type:varchar(100)"`
	Size        int64          `gorm:"type:bigint"`
	URL         string         `gorm:"type:varchar(500)"`
	ThumbURL    *string        `gorm:"column:thumb_url;type:varchar(500)"`
	Width       *int           `gorm:"type:int"`
	Height      *int           `gorm:"type:int"`
	StoragePath string         `gorm:"column:storage_path;type:varchar(500)"`
	StorageType string         `gorm:"column:storage_type;type:varchar(20);default:'local'"` // local/aliyun/tencent/minio
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (Media) TableName() string {
	return "media"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (m *Media) AfterFind(tx *gorm.DB) error {
	m.URL = resolveURL(m.URL)
	if m.ThumbURL != nil {
		resolved := resolveURL(*m.ThumbURL)
		m.ThumbURL = &resolved
	}
	return nil
}

// MediaModel 媒体文件模型操作结构体。
type MediaModel struct {
	db *gorm.DB
}

// NewMedia 创建 MediaModel 实例。
func NewMedia() *MediaModel {
	return &MediaModel{db: DB}
}

// Create 创建媒体文件记录。
func (m *MediaModel) Create(ctx context.Context, media *Media) error {
	return m.db.WithContext(ctx).Create(media).Error
}

// GetByID 根据 ID 查询媒体文件。
func (m *MediaModel) GetByID(ctx context.Context, id string) (*Media, error) {
	var media Media
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetList 分页查询媒体文件列表，支持按文件类型和文件名过滤，排除软删除记录。
func (m *MediaModel) GetList(ctx context.Context, fileType *int16, keyword *string, page, pageSize int) ([]Media, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Media{})

	if fileType != nil {
		query = query.Where("file_type = ?", *fileType)
	}
	if keyword != nil && *keyword != "" {
		query = query.Where("filename ILIKE ?", "%"+*keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var mediaList []Media
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&mediaList).Error
	if err != nil {
		return nil, 0, err
	}
	return mediaList, total, nil
}

// SoftDelete 软删除媒体文件。
func (m *MediaModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Media{}).Error
}

// GetByIDRaw 根据 ID 查询媒体文件，跳过 AfterFind 钩子，返回存储中的原始 URL。
// 用于引用扫描与删除：需要匹配的原始存储值而非拼接后的完整 URL。
func (m *MediaModel) GetByIDRaw(ctx context.Context, id string) (*Media, error) {
	var media Media
	err := m.db.WithContext(ctx).
		Session(&gorm.Session{SkipHooks: true}).
		Where("id = ?", id).
		First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// HardDeleteWithPresets 物理删除媒体及其全部预设记录（含软删记录）。
func (m *MediaModel) HardDeleteWithPresets(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("media_id = ?", id).Delete(&MediaPreset{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("id = ?", id).Delete(&Media{}).Error
	})
}

// GetByIDs 按 ID 列表批量查询媒体文件。
func (m *MediaModel) GetByIDs(ctx context.Context, ids []string) ([]Media, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var mediaList []Media
	err := m.db.WithContext(ctx).Where("id IN ?", ids).Find(&mediaList).Error
	if err != nil {
		return nil, err
	}
	return mediaList, nil
}

// GetAllIDs 查询所有未软删除的媒体文件 ID 列表。
func (m *MediaModel) GetAllIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := m.db.WithContext(ctx).Model(&Media{}).Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// UpdateStorageInfo 更新媒体文件的 URL 和存储类型（迁移完成后调用）。
func (m *MediaModel) UpdateStorageInfo(ctx context.Context, id, url, storageType string) error {
	return m.db.WithContext(ctx).Model(&Media{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"url":          url,
			"storage_type": storageType,
		}).Error
}

// CountNotOnStorage counts media files not on the target storage platform.
func (m *MediaModel) CountNotOnStorage(ctx context.Context, storageType string) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Media{}).Where("storage_type != ?", storageType).Count(&count).Error
	return count, err
}
