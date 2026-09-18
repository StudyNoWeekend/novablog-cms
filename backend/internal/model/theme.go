// Package model 定义数据库模型。
package model

import (
	"context"
	"time"
)

// Theme 已安装主题实例，对应 themes 数据表。
// 同一 theme_id 允许多版本共存，激活指针（bloggers.active_theme_id）引用具体实例。
type Theme struct {
	ID           string    `gorm:"type:uuid;primaryKey"`                                        // 实例 UUID 主键
	ThemeID      string    `gorm:"type:varchar(100);not null;uniqueIndex:uk_themes_id_version"` // theme.json 的 id
	Name         string    `gorm:"type:varchar(100);not null"`                                  // 展示名
	Version      string    `gorm:"type:varchar(50);not null;uniqueIndex:uk_themes_id_version"`  // 语义化版本
	Engine       string    `gorm:"type:varchar(50);not null"`                                   // 渲染引擎（next-static）
	APICompat    string    `gorm:"type:varchar(20);not null"`                                   // 兼容的公开 API 版本
	Author       string    `gorm:"type:varchar(100)"`                                           // 作者/组织
	Description  string    `gorm:"type:text"`                                                   // 一句话简介
	Screenshots  string    `gorm:"type:jsonb"`                                                  // 制品内截图相对路径 JSON 数组
	Fallbacks    string    `gorm:"type:jsonb"`                                                  // routes.fallback 壳页面映射 JSON
	Source       string    `gorm:"type:varchar(20);not null"`                                   // 来源：official | builtin
	SourceRef    string    `gorm:"type:varchar(500)"`                                           // 来源引用（owner/repo@tag 或 builtin 目录名）
	MarketID     int64     `gorm:"index"`                                                       // 官方市场主题 ID（检查更新用）
	MarketSlug   string    `gorm:"type:varchar(100)"`                                           // 官方市场主题 slug
	ArtifactPath string    `gorm:"type:varchar(500);not null"`                                  // 解压目录名（相对 themes.data_dir）
	Checksum     string    `gorm:"type:varchar(64)"`                                            // 制品 sha256 hex
	CreatedAt    time.Time `gorm:"type:timestamptz;autoCreateTime"`                             // 安装时间
	UpdatedAt    time.Time `gorm:"type:timestamptz;autoUpdateTime"`                             // 更新时间
}

// TableName 指定数据表名称。
func (Theme) TableName() string {
	return "themes"
}

// ThemeModel 主题模型操作结构体。
type ThemeModel struct{}

// NewTheme 创建 ThemeModel 实例。
func NewTheme() *ThemeModel {
	return &ThemeModel{}
}

// Create 创建主题安装记录。
func (m *ThemeModel) Create(ctx context.Context, theme *Theme) error {
	return DB.WithContext(ctx).Create(theme).Error
}

// GetByID 根据实例 UUID 查询主题。
func (m *ThemeModel) GetByID(ctx context.Context, id string) (*Theme, error) {
	var theme Theme
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&theme).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

// GetByThemeAndVersion 根据 theme_id + version 查询安装实例。
func (m *ThemeModel) GetByThemeAndVersion(ctx context.Context, themeID, version string) (*Theme, error) {
	var theme Theme
	if err := DB.WithContext(ctx).Where("theme_id = ? AND version = ?", themeID, version).First(&theme).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

// GetLatestByThemeID 查询指定 theme_id 最新的安装实例（预览用）。
func (m *ThemeModel) GetLatestByThemeID(ctx context.Context, themeID string) (*Theme, error) {
	var theme Theme
	if err := DB.WithContext(ctx).Where("theme_id = ?", themeID).Order("created_at DESC").First(&theme).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

// List 查询全部已安装主题（按安装时间倒序）。
func (m *ThemeModel) List(ctx context.Context) ([]Theme, error) {
	var themes []Theme
	if err := DB.WithContext(ctx).Order("created_at DESC").Find(&themes).Error; err != nil {
		return nil, err
	}
	return themes, nil
}

// Count 查询已安装主题数量。
func (m *ThemeModel) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := DB.WithContext(ctx).Model(&Theme{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Delete 物理删除安装记录（卸载）。
func (m *ThemeModel) Delete(ctx context.Context, id string) error {
	return DB.WithContext(ctx).Where("id = ?", id).Delete(&Theme{}).Error
}
