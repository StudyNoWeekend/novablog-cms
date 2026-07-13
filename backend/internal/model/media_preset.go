package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MediaPreset 媒体预设模型，对应 media_presets 数据表。
type MediaPreset struct {
	ID                string         `gorm:"type:uuid;primaryKey"`
	MediaID           string         `gorm:"column:media_id;type:uuid;not null"`
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

// SoftDelete 软删除媒体预设。
func (m *MediaPresetModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&MediaPreset{}).Error
}
