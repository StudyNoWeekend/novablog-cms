package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// OpenSourceWork 开源作品模型，对应 open_source_works 数据表。
// 展示技术人员维护的开源仓库：仓库名称、仓库链接、介绍（README 自动拉取）。
type OpenSourceWork struct {
	ID              string         `gorm:"type:uuid;primaryKey"`
	Name            string         `gorm:"type:varchar(255);not null"`                  // 仓库名称
	RepoURL         string         `gorm:"column:repo_url;type:varchar(1024);not null"` // 仓库链接
	Summary         string         `gorm:"type:varchar(500);default:''"`                // 一句话介绍（列表卡片展示）
	Readme          string         `gorm:"type:text"`                                   // README 原文（Markdown，自动拉取）
	Language        string         `gorm:"type:varchar(100);default:''"`                // 主语言
	Topics          string         `gorm:"type:varchar(500);default:''"`                // 主题标签，逗号分隔
	Stars           int            `gorm:"type:int;default:0"`                          // Star 数（刷新时快照）
	Homepage        string         `gorm:"column:homepage;type:varchar(1024)"`          // 主页/演示地址
	Status          int            `gorm:"type:int;default:0"`                          // 0=草稿, 1=已发布
	SortOrder       int            `gorm:"column:sort_order;type:int;default:0"`
	ReadmeUpdatedAt *time.Time     `gorm:"column:readme_updated_at;type:timestamptz"` // README 最近拉取时间
	CreatedAt       time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (OpenSourceWork) TableName() string {
	return "open_source_works"
}

// OpenSourceWorkModel 开源作品模型操作结构体。
type OpenSourceWorkModel struct {
	db *gorm.DB
}

// NewOpenSourceWork 创建 OpenSourceWorkModel 实例。
func NewOpenSourceWork() *OpenSourceWorkModel {
	return &OpenSourceWorkModel{db: DB}
}

// Create 创建开源作品记录。
func (m *OpenSourceWorkModel) Create(ctx context.Context, w *OpenSourceWork) error {
	return m.db.WithContext(ctx).Create(w).Error
}

// GetByID 根据 ID 查询开源作品。
func (m *OpenSourceWorkModel) GetByID(ctx context.Context, id string) (*OpenSourceWork, error) {
	var w OpenSourceWork
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// GetList 分页查询开源作品列表，支持按 keyword 模糊搜索名称/简介、按 status 过滤。
// 列表查询统一不返回 readme 大字段，需要 README 时用 GetByID。
func (m *OpenSourceWorkModel) GetList(ctx context.Context, keyword *string, status *int, page, pageSize int) ([]OpenSourceWork, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&OpenSourceWork{}).Omit("readme")

	if keyword != nil && *keyword != "" {
		like := "%" + *keyword + "%"
		query = query.Where("name ILIKE ? OR summary ILIKE ?", like, like)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []OpenSourceWork
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, stars DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新开源作品。
func (m *OpenSourceWorkModel) Update(ctx context.Context, w *OpenSourceWork) error {
	return m.db.WithContext(ctx).Save(w).Error
}

// SoftDelete 软删除开源作品。
func (m *OpenSourceWorkModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&OpenSourceWork{}).Error
}
