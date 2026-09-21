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
	CategoryID   *string        `gorm:"column:category_id;type:uuid;index:idx_articles_category_id"`
	Category     *Category      `gorm:"foreignKey:CategoryID"`
	Status       int16          `gorm:"column:status;type:smallint;not null;default:1;index:idx_articles_status"`
	Type         int16          `gorm:"type:smallint;default:1"`
	Extra        map[string]any `gorm:"type:jsonb;serializer:json"`
	ViewCount    int            `gorm:"type:int;default:0;column:view_count"`
	CommentCount int            `gorm:"type:int;default:0;column:comment_count"`
	IsTop        bool           `gorm:"type:boolean;default:false;column:is_top"`
	IsComment    bool           `gorm:"type:boolean;default:true;column:is_comment"`
	PublishedAt  *time.Time     `gorm:"column:published_at;type:timestamptz;index:idx_articles_published_at"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Tags         []Tag          `gorm:"many2many:article_tags"`
}

// TableName 指定数据表名称。
func (Article) TableName() string {
	return "articles"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (a *Article) AfterFind(tx *gorm.DB) error {
	a.CoverImage = resolveURL(a.CoverImage)
	return nil
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

// CreateWithTx creates an article within a transaction.
func (m *ArticleModel) CreateWithTx(ctx context.Context, tx *gorm.DB, article *Article) error {
	return tx.WithContext(ctx).Create(article).Error
}

// UpdateWithTx updates an article within a transaction.
func (m *ArticleModel) UpdateWithTx(ctx context.Context, tx *gorm.DB, article *Article) error {
	return tx.WithContext(ctx).Save(article).Error
}

// ReplaceTagsWithTx replaces tags within a transaction.
func (m *ArticleModel) ReplaceTagsWithTx(ctx context.Context, tx *gorm.DB, articleID string, tagIDs []string) error {
	// Delete existing associations
	if err := tx.WithContext(ctx).Where("article_id = ?", articleID).Delete(&ArticleTag{}).Error; err != nil {
		return err
	}
	// Create new associations
	if len(tagIDs) == 0 {
		return nil
	}
	associations := make([]ArticleTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		associations = append(associations, ArticleTag{
			ArticleID: articleID,
			TagID:     tagID,
		})
	}
	return tx.WithContext(ctx).Create(&associations).Error
}

// Transaction wraps a function in a DB transaction.
func (m *ArticleModel) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
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

// GetHotList 查询热门文章列表，按浏览量降序取指定数量。
func (m *ArticleModel) GetHotList(ctx context.Context, count int) ([]Article, error) {
	var articles []Article
	err := m.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Where("status = ?", 2).
		Order("view_count DESC").
		Limit(count).
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	return articles, nil
}

// GetRandomList 随机查询已发布文章列表，取指定数量。
// 使用 ORDER BY RANDOM() 配合 LIMIT，适用于中小规模数据集。
// 对于大数据集，可考虑 TABLESAMPLE BERNOULLI(percentage) LIMIT ? 等更高效方案。
func (m *ArticleModel) GetRandomList(ctx context.Context, count int) ([]Article, error) {
	var articles []Article
	err := m.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Where("status = ?", 2).
		Order("RANDOM()").
		Limit(count).
		Find(&articles).Error
	if err != nil {
		return nil, err
	}
	return articles, nil
}

// IncrementViewCount 根据slug增加文章浏览量，仅对已发布文章生效。
func (m *ArticleModel) IncrementViewCount(ctx context.Context, slug string) error {
	return m.db.WithContext(ctx).
		Model(&Article{}).
		Where("slug = ? AND status = ?", slug, 2).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementViewCountByID 根据 ID 增加文章浏览量，仅对已发布文章生效。
func (m *ArticleModel) IncrementViewCountByID(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).
		Model(&Article{}).
		Where("id = ? AND status = ?", id, 2).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateCommentCount 增减文章评论计数，仅对已发布文章生效。delta 为 +1（新增）或 -1（删除）。
func (m *ArticleModel) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	expr := gorm.Expr("comment_count + ?", delta)
	if delta < 0 {
		expr = gorm.Expr("GREATEST(comment_count - ?, 0)", -delta)
	}
	return m.db.WithContext(ctx).
		Model(&Article{}).
		Where("id = ? AND status = ?", id, 2).
		UpdateColumn("comment_count", expr).Error
}
