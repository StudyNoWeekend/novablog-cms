// Package model 定义数据库模型。
package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DB 全局数据库连接实例。
var DB *gorm.DB

// BaseURL 本地文件访问基础URL，由 bootstrap 阶段注入。
// AfterFind Hook 使用此变量将相对路径拼接为完整 URL。
var BaseURL string

// Blogger 博主信息模型，对应 bloggers 数据表。
type Blogger struct {
	ID              string         `gorm:"type:uuid;primaryKey"`                  // UUID 主键
	Username        string         `gorm:"type:varchar(50);uniqueIndex;not null"` // 用户名，唯一
	PasswordHash    string         `gorm:"type:varchar(255);not null"`            // 密码哈希值
	Nickname        string         `gorm:"type:varchar(50)"`                      // 昵称
	Avatar          string         `gorm:"type:varchar(500)"`                     // 头像 URL
	Bio             string         `gorm:"type:text"`                             // 个人简介
	Email           string         `gorm:"type:varchar(100)"`                     // 邮箱
	City            string         `gorm:"type:varchar(100)"`                     // 所在城市（省/市）
	BlogTitle       string         `gorm:"type:varchar(100)"`                     // 博客标题
	BlogDescription string         `gorm:"type:text"`                             // 博客描述
	BlogIcon        string         `gorm:"type:varchar(500)"`                     // 博客 icon 图 URL
	PageBackground  string         `gorm:"type:varchar(500)"`                     // 页面背景图 URL
	SocialLinks     string         `gorm:"type:jsonb"`                            // 社交平台链接 JSON 数组
	Tags            string         `gorm:"type:jsonb"`                            // 标签 JSON 字符串数组
	ActiveThemeID   *string        `gorm:"type:uuid;index"`                       // 激活主题实例 ID（themes.id，nil=未激活）
	LastLoginAt     *time.Time     `gorm:"type:timestamptz"`                      // 最后登录时间
	CreatedAt       time.Time      `gorm:"type:timestamptz;autoCreateTime"`       // 创建时间
	UpdatedAt       time.Time      `gorm:"type:timestamptz;autoUpdateTime"`       // 更新时间
	DeletedAt       gorm.DeletedAt `gorm:"index"`                                 // 软删除时间
}

// TableName 指定数据表名称。
func (Blogger) TableName() string {
	return "bloggers"
}

// BeforeCreate GORM 创建前钩子：确保 JSONB 字段为合法 JSON，避免空字符串触发 PostgreSQL 22P02。
func (b *Blogger) BeforeCreate(tx *gorm.DB) error {
	if b.SocialLinks == "" {
		b.SocialLinks = "[]"
	}
	if b.Tags == "" {
		b.Tags = "[]"
	}
	return nil
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (b *Blogger) AfterFind(tx *gorm.DB) error {
	b.Avatar = resolveURL(b.Avatar)
	b.BlogIcon = resolveURL(b.BlogIcon)
	b.PageBackground = resolveURL(b.PageBackground)
	return nil
}

// BloggerModel 博主模型操作结构体。
type BloggerModel struct {
	db *gorm.DB
}

// NewBlogger 创建 BloggerModel 实例。
func NewBlogger() *BloggerModel {
	return &BloggerModel{db: DB}
}

// GetByUsername 根据用户名查询博主信息。
func (m *BloggerModel) GetByUsername(ctx context.Context, username string) (*Blogger, error) {
	var blogger Blogger
	err := m.db.WithContext(ctx).Where("username = ?", username).First(&blogger).Error
	if err != nil {
		return nil, err
	}
	return &blogger, nil
}

// GetByID 根据 ID 查询博主信息。
func (m *BloggerModel) GetByID(ctx context.Context, id string) (*Blogger, error) {
	var blogger Blogger
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&blogger).Error
	if err != nil {
		return nil, err
	}
	return &blogger, nil
}

// Update 更新博主信息。
func (m *BloggerModel) Update(ctx context.Context, blogger *Blogger) error {
	return m.db.WithContext(ctx).Save(blogger).Error
}

// UpdateLastLogin 更新博主的最后登录时间。
func (m *BloggerModel) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return m.db.WithContext(ctx).Model(&Blogger{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// Count 查询博主数量。
func (m *BloggerModel) Count(ctx context.Context) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Blogger{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Create 创建博主记录。
func (m *BloggerModel) Create(ctx context.Context, blogger *Blogger) error {
	return m.db.WithContext(ctx).Create(blogger).Error
}

// GetFirst 查询第一条博主记录。
func (m *BloggerModel) GetFirst(ctx context.Context) (*Blogger, error) {
	var blogger Blogger
	err := m.db.WithContext(ctx).First(&blogger).Error
	if err != nil {
		return nil, err
	}
	return &blogger, nil
}

// UpdateActiveTheme 更新激活主题指针（themeID 传空串表示取消激活）。
func (m *BloggerModel) UpdateActiveTheme(ctx context.Context, id, themeID string) error {
	var ptr *string
	if themeID != "" {
		ptr = &themeID
	}
	return m.db.WithContext(ctx).Model(&Blogger{}).Where("id = ?", id).Update("active_theme_id", ptr).Error
}
