package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// VideoWork 视频作品模型，对应 video_works 数据表。
type VideoWork struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Title       string         `gorm:"type:varchar(255);not null"`
	CoverURL    string         `gorm:"column:cover_url;type:varchar(1024)"`
	Description string         `gorm:"type:text"`
	Status      int            `gorm:"type:int;default:0"` // 0=草稿, 1=已发布
	SortOrder   int            `gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (VideoWork) TableName() string {
	return "video_works"
}

// VideoWorkModel 视频作品模型操作结构体。
type VideoWorkModel struct {
	db *gorm.DB
}

// NewVideoWork 创建 VideoWorkModel 实例。
func NewVideoWork() *VideoWorkModel {
	return &VideoWorkModel{db: DB}
}

// Create 创建视频作品记录。
func (m *VideoWorkModel) Create(ctx context.Context, video *VideoWork) error {
	return m.db.WithContext(ctx).Create(video).Error
}

// GetByID 根据 ID 查询视频作品。
func (m *VideoWorkModel) GetByID(ctx context.Context, id string) (*VideoWork, error) {
	var video VideoWork
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetList 分页查询视频作品列表，支持按 keyword 模糊搜索标题、按 status 过滤。
func (m *VideoWorkModel) GetList(ctx context.Context, keyword *string, status *int, page, pageSize int) ([]VideoWork, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&VideoWork{})

	if keyword != nil && *keyword != "" {
		query = query.Where("title ILIKE ?", "%"+*keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []VideoWork
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

// Update 更新视频作品。
func (m *VideoWorkModel) Update(ctx context.Context, video *VideoWork) error {
	return m.db.WithContext(ctx).Save(video).Error
}

// SoftDelete 软删除视频作品。
func (m *VideoWorkModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&VideoWork{}).Error
}

// Transaction 暴露事务能力，供 logic 层使用。
func (m *VideoWorkModel) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}

// VideoPlatformLink 视频平台链接模型，对应 video_platform_links 数据表。
type VideoPlatformLink struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	VideoID   string         `gorm:"column:video_id;type:uuid;not null"`
	Platform  string         `gorm:"type:varchar(50);not null"`
	URL       string         `gorm:"column:url;type:varchar(1024);not null"`
	CreatedAt time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (VideoPlatformLink) TableName() string {
	return "video_platform_links"
}

// VideoPlatformLinkModel 视频平台链接模型操作结构体。
type VideoPlatformLinkModel struct {
	db *gorm.DB
}

// NewVideoPlatformLink 创建 VideoPlatformLinkModel 实例。
func NewVideoPlatformLink() *VideoPlatformLinkModel {
	return &VideoPlatformLinkModel{db: DB}
}

// BatchCreate 批量创建平台链接记录，使用传入的事务。
func (m *VideoPlatformLinkModel) BatchCreate(ctx context.Context, tx *gorm.DB, links []VideoPlatformLink) error {
	return tx.WithContext(ctx).Create(&links).Error
}

// GetByVideoID 查询指定视频作品下的所有平台链接，按 created_at 升序。
func (m *VideoPlatformLinkModel) GetByVideoID(ctx context.Context, videoID string) ([]VideoPlatformLink, error) {
	var links []VideoPlatformLink
	err := m.db.WithContext(ctx).
		Where("video_id = ? AND deleted_at IS NULL", videoID).
		Order("created_at ASC").
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// GetByVideoIDs 批量查询多个视频作品的平台链接，供列表页使用。
func (m *VideoPlatformLinkModel) GetByVideoIDs(ctx context.Context, videoIDs []string) ([]VideoPlatformLink, error) {
	var links []VideoPlatformLink
	if len(videoIDs) == 0 {
		return links, nil
	}
	err := m.db.WithContext(ctx).
		Where("video_id IN ? AND deleted_at IS NULL", videoIDs).
		Order("created_at ASC").
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// SoftDeleteByVideoID 按视频作品 ID 级联软删除所有平台链接，使用传入的事务。
func (m *VideoPlatformLinkModel) SoftDeleteByVideoID(ctx context.Context, tx *gorm.DB, videoID string) error {
	return tx.WithContext(ctx).
		Where("video_id = ?", videoID).
		Delete(&VideoPlatformLink{}).Error
}
