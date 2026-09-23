package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MediaPreset 媒体预设模型，对应 media_presets 数据表。
type MediaPreset struct {
	ID                string         `gorm:"type:uuid;primaryKey"`
	MediaID           string         `gorm:"column:media_id;type:uuid;not null;index:idx_media_presets_media_id"`
	Name              string         `gorm:"type:varchar(255);not null"`
	FrameConfig       string         `gorm:"column:frame_config;type:jsonb"`
	DisplayParams     string         `gorm:"column:display_params;type:jsonb"`
	OutputURL         string         `gorm:"column:output_url;type:varchar(500);not null"`
	OutputStoragePath string         `gorm:"column:output_storage_path;type:varchar(500);not null"`
	OutputSize        int64          `gorm:"column:output_size;type:bigint"`
	MimeType          string         `gorm:"column:mime_type;type:varchar(100)"`
	CreatedAt         time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (MediaPreset) TableName() string {
	return "media_presets"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (m *MediaPreset) AfterFind(tx *gorm.DB) error {
	m.OutputURL = resolveURL(m.OutputURL)
	return nil
}

// MediaPresetModel 媒体预设模型操作结构体。
type MediaPresetModel struct {
	db *gorm.DB
}

// NewMediaPreset 创建 MediaPresetModel 实例。
func NewMediaPreset() *MediaPresetModel {
	return &MediaPresetModel{db: DB}
}

// Create 创建媒体预设记录。
func (m *MediaPresetModel) Create(ctx context.Context, preset *MediaPreset) error {
	return m.db.WithContext(ctx).Create(preset).Error
}

// GetByID 根据 ID 查询媒体预设。
func (m *MediaPresetModel) GetByID(ctx context.Context, id string) (*MediaPreset, error) {
	var preset MediaPreset
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&preset).Error
	if err != nil {
		return nil, err
	}
	return &preset, nil
}

// GetByIDs batch queries presets by IDs.
func (m *MediaPresetModel) GetByIDs(ctx context.Context, ids []string) ([]MediaPreset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var presets []MediaPreset
	err := m.db.WithContext(ctx).Where("id IN ?", ids).Find(&presets).Error
	return presets, err
}

// GetByMediaID 查询指定原图下的所有未删除预设，按创建时间倒序。
func (m *MediaPresetModel) GetByMediaID(ctx context.Context, mediaID string) ([]MediaPreset, error) {
	var presets []MediaPreset
	err := m.db.WithContext(ctx).
		Where("media_id = ?", mediaID).
		Order("created_at DESC").
		Find(&presets).Error
	if err != nil {
		return nil, err
	}
	return presets, nil
}

// GetAllByMediaIDUnscoped 查询指定原图下的全部预设（含软删除），用于删除清理与引用扫描。
func (m *MediaPresetModel) GetAllByMediaIDUnscoped(ctx context.Context, mediaID string) ([]MediaPreset, error) {
	var presets []MediaPreset
	err := m.db.WithContext(ctx).Unscoped().
		Where("media_id = ?", mediaID).
		Order("created_at DESC").
		Find(&presets).Error
	return presets, err
}

// SoftDelete 软删除媒体预设。
func (m *MediaPresetModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&MediaPreset{}).Error
}

// GetAll returns all media presets with pagination.
func (m *MediaPresetModel) GetAll(ctx context.Context, page, pageSize int) ([]MediaPreset, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&MediaPreset{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var presets []MediaPreset
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&presets).Error
	if err != nil {
		return nil, 0, err
	}
	return presets, total, nil
}

// UpdateStorageInfo updates output_url for a preset.
func (m *MediaPresetModel) UpdateStorageInfo(ctx context.Context, id, outputURL string) error {
	return m.db.WithContext(ctx).Model(&MediaPreset{}).
		Where("id = ?", id).
		Update("output_url", outputURL).Error
}
