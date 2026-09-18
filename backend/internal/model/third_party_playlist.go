package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ThirdPartyPlaylist 第三方歌单模型，对应 third_party_playlists 数据表。
type ThirdPartyPlaylist struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(500);not null" json:"title"`
	CoverURL    string    `gorm:"type:text;column:cover_url" json:"cover_url"`
	Platform    string    `gorm:"type:varchar(50);not null" json:"platform"`
	PlatformURL string    `gorm:"type:text;column:platform_url;not null" json:"platform_url"`
	Description string    `gorm:"type:text" json:"description"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName 指定数据表名称。
func (ThirdPartyPlaylist) TableName() string {
	return "third_party_playlists"
}

// ThirdPartyPlaylistModel 第三方歌单模型操作结构体。
type ThirdPartyPlaylistModel struct {
	db *gorm.DB
}

// NewThirdPartyPlaylist 创建 ThirdPartyPlaylistModel 实例。
func NewThirdPartyPlaylist() *ThirdPartyPlaylistModel {
	return &ThirdPartyPlaylistModel{db: DB}
}

// Create 创建歌单。
func (m *ThirdPartyPlaylistModel) Create(ctx context.Context, playlist *ThirdPartyPlaylist) error {
	return m.db.WithContext(ctx).Create(playlist).Error
}

// GetByID 根据 ID 查询歌单。
func (m *ThirdPartyPlaylistModel) GetByID(ctx context.Context, id string) (*ThirdPartyPlaylist, error) {
	var playlist ThirdPartyPlaylist
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&playlist).Error
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// GetList 分页查询歌单列表，按 sort_order ASC, created_at DESC 排序。
func (m *ThirdPartyPlaylistModel) GetList(ctx context.Context, page, pageSize int) (playlists []ThirdPartyPlaylist, total int64, err error) {
	query := m.db.WithContext(ctx).Model(&ThirdPartyPlaylist{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&playlists).Error
	if err != nil {
		return nil, 0, err
	}
	return playlists, total, nil
}

// GetPublicList 查询前台展示的歌单列表（仅 enabled=true），按 sort_order ASC, created_at DESC 排序。
func (m *ThirdPartyPlaylistModel) GetPublicList(ctx context.Context) (playlists []ThirdPartyPlaylist, err error) {
	err = m.db.WithContext(ctx).
		Model(&ThirdPartyPlaylist{}).
		Where("enabled = ?", true).
		Order("sort_order ASC, created_at DESC").
		Find(&playlists).Error
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

// Update 更新歌单。
func (m *ThirdPartyPlaylistModel) Update(ctx context.Context, playlist *ThirdPartyPlaylist) error {
	return m.db.WithContext(ctx).Save(playlist).Error
}

// Delete 删除歌单。
func (m *ThirdPartyPlaylistModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&ThirdPartyPlaylist{}).Error
}
