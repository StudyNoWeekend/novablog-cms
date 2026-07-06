package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Article 文章模型，对应 articles 数据表。
type Article struct {
	ID           string         `gorm:"type:uuid;primaryKey"`
	Title        string         `gorm:"type:varchar(200);not null"`
	Slug         string         `gorm:"type:varchar(200);uniqueIndex"`
	Summary      string         `gorm:"type:varchar(500)"`
	Content      string         `gorm:"type:text;not null"`
	CoverImage   string         `gorm:"type:varchar(500);column:cover_image"`
	CategoryID   *string        `gorm:"type:uuid;column:category_id"`
	Category     *Category      `gorm:"foreignKey:CategoryID"`
	Status       int16          `gorm:"type:smallint;default:1"`
	Type         int16          `gorm:"type:smallint;default:1"`
	Extra        map[string]any `gorm:"type:jsonb;serializer:json"`
	ViewCount    int            `gorm:"type:int;default:0;column:view_count"`
	CommentCount int            `gorm:"type:int;default:0;column:comment_count"`
	IsTop        bool           `gorm:"type:boolean;default:false;column:is_top"`
	IsComment    bool           `gorm:"type:boolean;default:true;column:is_comment"`
	PublishedAt  *time.Time     `gorm:"type:timestamptz;column:published_at"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Tags         []Tag          `gorm:"many2many:article_tags"`
}

// TableName 指定数据表名称。
func (Article) TableName() string {
	return "articles"
}

// ArticleModel 文章模型操作结构体。
type ArticleModel struct {
	db *gorm.DB
}

// NewArticle 创建 ArticleModel 实例。
func NewArticle() *ArticleModel {
	return &ArticleModel{db: DB}
}

// Create 创建文章。
func (m *ArticleModel) Create(ctx context.Context, article *Article) error {
	return m.db.WithContext(ctx).Create(article).Error
}

// GetByID 根据 ID 查询文章，包含 Category 和 Tags 预加载。
func (m *ArticleModel) GetByID(ctx context.Context, id string) (*Article, error) {
	var article Article
	err := m.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Where("id = ?", id).
		First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetBySlug 根据 slug 查询文章，包含 Category 和 Tags 预加载。
func (m *ArticleModel) GetBySlug(ctx context.Context, slug string) (*Article, error) {
	var article Article
	err := m.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Where("slug = ?", slug).
		First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetList 分页查询文章列表，支持按 status、category_id 筛选，keyword 模糊搜索 title。
// 排除软删除记录，预加载 Category 和 Tags，按 is_top DESC, published_at DESC, created_at DESC 排序。
func (m *ArticleModel) GetList(ctx context.Context, page, pageSize int, status *int16, categoryID *string, keyword *string) (articles []Article, total int64, err error) {
	query := m.db.WithContext(ctx).Model(&Article{}).Preload("Category").Preload("Tags")

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if keyword != nil && *keyword != "" {
		query = query.Where("title LIKE ?", "%"+*keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.
		Order("is_top DESC, published_at DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

// Update 更新文章。
func (m *ArticleModel) Update(ctx context.Context, article *Article) error {
	return m.db.WithContext(ctx).Save(article).Error
}

// UpdateStatus 更新文章状态，如果 status=2 则设置 published_at。
func (m *ArticleModel) UpdateStatus(ctx context.Context, id string, status int16) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == 2 {
		now := time.Now()
		updates["published_at"] = &now
	}
	return m.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).Updates(updates).Error
}

// SoftDelete 软删除文章。
func (m *ArticleModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Article{}).Error
}

// ExistsBySlug 检查 slug 是否已存在，可排除指定 ID。
func (m *ArticleModel) ExistsBySlug(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int64
	query := m.db.WithContext(ctx).Model(&Article{}).Where("slug = ?", slug)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
