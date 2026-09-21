package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Project 项目经历模型，对应 projects 数据表。
// 通用项目经历展示，覆盖摄影、视频剪辑、技术开发等任意行业。
type Project struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Category    string         `gorm:"type:varchar(100);default:''"` // 领域分类：摄影/视频剪辑/技术开发/设计等
	Role        string         `gorm:"type:varchar(100);default:''"` // 担任角色
	Client      string         `gorm:"type:varchar(255);default:''"` // 客户/所属组织
	CoverURL    string         `gorm:"column:cover_url;type:varchar(1024)"`
	Summary     string         `gorm:"type:varchar(500);default:''"` // 一句话简介
	Description string         `gorm:"type:text"`
	TechStack   string         `gorm:"type:varchar(500);default:''"` // 技能/工具标签，逗号分隔
	StartDate   *time.Time     `gorm:"type:date"` // 开始时间（按月粒度，存当月 1 号）
	EndDate     *time.Time     `gorm:"type:date"` // 结束时间，null 表示至今
	ProjectURL  string         `gorm:"column:project_url;type:varchar(1024)"`
	RepoURL     string         `gorm:"column:repo_url;type:varchar(1024)"`
	Status      int            `gorm:"type:int;default:0"` // 0=草稿, 1=已发布
	SortOrder   int            `gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (Project) TableName() string {
	return "projects"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (p *Project) AfterFind(tx *gorm.DB) error {
	p.CoverURL = resolveURL(p.CoverURL)
	return nil
}

// ProjectModel 项目经历模型操作结构体。
type ProjectModel struct {
	db *gorm.DB
}

// NewProject 创建 ProjectModel 实例。
func NewProject() *ProjectModel {
	return &ProjectModel{db: DB}
}

// Create 创建项目经历记录。
func (m *ProjectModel) Create(ctx context.Context, p *Project) error {
	return m.db.WithContext(ctx).Create(p).Error
}

// GetByID 根据 ID 查询项目经历。
func (m *ProjectModel) GetByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetList 分页查询项目经历列表，支持按 keyword 模糊搜索标题/简介、按 category 过滤、按 status 过滤。
func (m *ProjectModel) GetList(ctx context.Context, keyword, category *string, status *int, page, pageSize int) ([]Project, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Project{})

	if keyword != nil && *keyword != "" {
		like := "%" + *keyword + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ?", like, like)
	}
	if category != nil && *category != "" {
		query = query.Where("category = ?", *category)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Project
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, start_date DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新项目经历。
func (m *ProjectModel) Update(ctx context.Context, p *Project) error {
	return m.db.WithContext(ctx).Save(p).Error
}

// SoftDelete 软删除项目经历。
func (m *ProjectModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Project{}).Error
}
