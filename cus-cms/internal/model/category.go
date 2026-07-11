package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Category 分类模型，对应 categories 数据表。
type Category struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Slug        string    `gorm:"type:varchar(100);uniqueIndex"`
	Description string    `gorm:"type:varchar(500)"`
	Type        string    `gorm:"type:varchar(50);default:article" json:"type"`
	SortOrder   int       `gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time `gorm:"type:timestamptz;autoCreateTime"`
}

// TableName 指定数据表名称。
func (Category) TableName() string {
	return "categories"
}

// CategoryModel 分类模型操作结构体。
type CategoryModel struct {
	db *gorm.DB
}

// NewCategory 创建 CategoryModel 实例。
func NewCategory() *CategoryModel {
	return &CategoryModel{db: DB}
}

// Create 创建分类。
func (m *CategoryModel) Create(ctx context.Context, category *Category) error {
	return m.db.WithContext(ctx).Create(category).Error
}

// GetByID 根据 ID 查询分类。
func (m *CategoryModel) GetByID(ctx context.Context, id string) (*Category, error) {
	var category Category
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByName 根据名称查询分类，可按 type 过滤。
func (m *CategoryModel) GetByName(ctx context.Context, name string, categoryType ...string) (*Category, error) {
	var category Category
	query := m.db.WithContext(ctx).Where("name = ?", name)
	if len(categoryType) > 0 && categoryType[0] != "" {
		query = query.Where("type = ?", categoryType[0])
	}
	err := query.First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetAll 获取所有分类，按 sort_order 排序，可按 type 过滤。
func (m *CategoryModel) GetAll(ctx context.Context, categoryType ...string) ([]Category, error) {
	var categories []Category
	query := m.db.WithContext(ctx)
	if len(categoryType) > 0 && categoryType[0] != "" {
		query = query.Where("type = ?", categoryType[0])
	}
	err := query.Order("sort_order").Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// Update 更新分类。
func (m *CategoryModel) Update(ctx context.Context, category *Category) error {
	return m.db.WithContext(ctx).Save(category).Error
}

// Delete 删除分类。
func (m *CategoryModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Category{}).Error
}

// HasArticles 检查分类下是否有文章。
func (m *CategoryModel) HasArticles(ctx context.Context, id string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Article{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasSongs 检查分类下是否有歌曲。
func (m *CategoryModel) HasSongs(ctx context.Context, id string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Song{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasPortfolios 检查分类下是否有作品集。
func (m *CategoryModel) HasPortfolios(ctx context.Context, id string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&Portfolio{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasTravelGuides 检查分类下是否有旅行攻略。
func (m *CategoryModel) HasTravelGuides(ctx context.Context, id string) (bool, error) {
	var count int64
	err := m.db.WithContext(ctx).Model(&TravelGuide{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
