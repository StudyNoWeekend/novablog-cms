package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Song 歌曲模型，对应 songs 数据表。
type Song struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title      string    `gorm:"type:varchar(500);not null" json:"title"`
	Artist     string    `gorm:"type:varchar(255);not null" json:"artist"`
	CoverURL   string    `gorm:"type:text;column:cover_url" json:"cover_url"`
	BVID       string    `gorm:"type:varchar(50);column:bvid;not null" json:"bvid"`
	CID        int64     `gorm:"column:cid;not null" json:"cid"`
	SourceURL  string    `gorm:"type:text;column:source_url;not null" json:"source_url"`
	SourceType string    `gorm:"type:varchar(50);column:source_type;default:bilibili" json:"source_type"`
	CategoryID *string   `gorm:"type:uuid;column:category_id" json:"category_id"`
	Duration   int       `gorm:"default:0" json:"duration"`
	SortOrder  int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName 指定数据表名称。
func (Song) TableName() string {
	return "songs"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (s *Song) AfterFind(tx *gorm.DB) error {
	s.CoverURL = resolveURL(s.CoverURL)
	return nil
}

// SongModel 歌曲模型操作结构体。
type SongModel struct {
	db *gorm.DB
}

// NewSong 创建 SongModel 实例。
func NewSong() *SongModel {
	return &SongModel{db: DB}
}

// Create 创建歌曲。
func (m *SongModel) Create(ctx context.Context, song *Song) error {
	return m.db.WithContext(ctx).Create(song).Error
}

// GetByID 根据 ID 查询歌曲。
func (m *SongModel) GetByID(ctx context.Context, id string) (*Song, error) {
	var song Song
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&song).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// GetByIDRaw 根据 ID 查询歌曲，跳过 AfterFind 钩子，返回存储中的原始 cover_url。
// 用于更新路径：避免把钩子解析出的完整 URL 原样写回，覆盖相对路径存储值。
func (m *SongModel) GetByIDRaw(ctx context.Context, id string) (*Song, error) {
	var song Song
	err := m.db.WithContext(ctx).
		Session(&gorm.Session{SkipHooks: true}).
		Where("id = ?", id).
		First(&song).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// GetList 分页查询歌曲列表，支持按 category_id 筛选。
// 按 sort_order ASC, created_at DESC 排序。
func (m *SongModel) GetList(ctx context.Context, categoryID *string, page, pageSize int) (songs []Song, total int64, err error) {
	query := m.db.WithContext(ctx).Model(&Song{})

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.
		Order("sort_order ASC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&songs).Error
	if err != nil {
		return nil, 0, err
	}
	return songs, total, nil
}

// Update 更新歌曲。
func (m *SongModel) Update(ctx context.Context, song *Song) error {
	return m.db.WithContext(ctx).Save(song).Error
}

// Delete 删除歌曲。
func (m *SongModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Song{}).Error
}
