package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Recipe 美食菜谱模型，对应 recipes 数据表。
type Recipe struct {
	ID          string           `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string           `gorm:"type:varchar(255);not null" json:"title"`
	Cover       string           `gorm:"type:varchar(500)" json:"cover"`
	Summary     string           `gorm:"type:varchar(500)" json:"summary"`
	Ingredients []map[string]any `gorm:"type:jsonb;serializer:json" json:"ingredients"` // 食材清单 [{name,amount}]
	Steps       []map[string]any `gorm:"type:jsonb;serializer:json" json:"steps"`       // 步骤 [{text,image}]
	Difficulty  int              `gorm:"type:int;default:1" json:"difficulty"`          // 1=简单, 2=中等, 3=困难
	Minutes     int              `gorm:"type:int;default:0" json:"minutes"`             // 总耗时（分钟）
	Servings    int              `gorm:"type:int;default:1" json:"servings"`            // 份量
	Tags        string           `gorm:"type:varchar(500);default:''" json:"tags"`      // 标签，逗号分隔
	Status      int              `gorm:"type:int;default:0" json:"status"`              // 0=草稿, 1=已发布
	SortOrder   int              `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	ViewCount   int              `gorm:"column:view_count;type:int;default:0" json:"view_count"`
	CreatedAt   time.Time        `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定数据表名称。
func (Recipe) TableName() string {
	return "recipes"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (r *Recipe) AfterFind(tx *gorm.DB) error {
	r.Cover = resolveURL(r.Cover)
	return nil
}

// RecipeModel 美食菜谱模型操作结构体。
type RecipeModel struct {
	db *gorm.DB
}

// NewRecipe 创建 RecipeModel 实例。
func NewRecipe() *RecipeModel {
	return &RecipeModel{db: DB}
}

// Create 创建菜谱记录。
func (m *RecipeModel) Create(ctx context.Context, r *Recipe) error {
	return m.db.WithContext(ctx).Create(r).Error
}

// GetByID 根据 ID 查询菜谱。
func (m *RecipeModel) GetByID(ctx context.Context, id string) (*Recipe, error) {
	var r Recipe
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetList 分页查询菜谱列表，支持 keyword 模糊搜索标题/摘要/标签、difficulty 筛选、status 过滤。
func (m *RecipeModel) GetList(ctx context.Context, keyword *string, difficulty *int, status *int, page, pageSize int) ([]Recipe, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Recipe{})

	if keyword != nil && *keyword != "" {
		like := "%" + *keyword + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ? OR tags ILIKE ?", like, like, like)
	}
	if difficulty != nil && *difficulty > 0 {
		query = query.Where("difficulty = ?", *difficulty)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Recipe
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

// Update 更新菜谱。
func (m *RecipeModel) Update(ctx context.Context, r *Recipe) error {
	return m.db.WithContext(ctx).Save(r).Error
}

// SoftDelete 软删除菜谱。
func (m *RecipeModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Recipe{}).Error
}
