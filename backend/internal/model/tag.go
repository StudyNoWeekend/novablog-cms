package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Tag 标签模型，对应 tags 数据表。
type Tag struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime"`
}

// TableName 指定数据表名称。
func (Tag) TableName() string {
	return "tags"
}

// TagModel 标签模型操作结构体。
type TagModel struct {
	db *gorm.DB
}

// NewTag 创建 TagModel 实例。
func NewTag() *TagModel {
	return &TagModel{db: DB}
}

// Create 创建标签。
func (m *TagModel) Create(ctx context.Context, tag *Tag) error {
	return m.db.WithContext(ctx).Create(tag).Error
}

// GetByID 根据 ID 查询标签。
func (m *TagModel) GetByID(ctx context.Context, id string) (*Tag, error) {
	var tag Tag
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称查询标签。
func (m *TagModel) GetByName(ctx context.Context, name string) (*Tag, error) {
	var tag Tag
	err := m.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetAll 获取所有标签。
func (m *TagModel) GetAll(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := m.db.WithContext(ctx).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// Update 更新标签。
func (m *TagModel) Update(ctx context.Context, tag *Tag) error {
	return m.db.WithContext(ctx).Save(tag).Error
}

// Delete 删除标签。
func (m *TagModel) Delete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Tag{}).Error
}

// ArticleTag 文章标签关联表模型，对应 article_tags 数据表。
type ArticleTag struct {
	ArticleID string `gorm:"type:uuid;column:article_id;primaryKey"`
	TagID     string `gorm:"type:uuid;column:tag_id;primaryKey"`
}

// TableName 指定数据表名称。
func (ArticleTag) TableName() string {
	return "article_tags"
}

// ArticleTagModel 文章标签关联表操作结构体。
type ArticleTagModel struct {
	db *gorm.DB
}

// NewArticleTag 创建 ArticleTagModel 实例。
func NewArticleTag() *ArticleTagModel {
	return &ArticleTagModel{db: DB}
}

// ReplaceTags 替换文章标签关联：先删除旧关联，再批量插入新关联，使用事务。
func (m *ArticleTagModel) ReplaceTags(ctx context.Context, articleID string, tagIDs []string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", articleID).Delete(&ArticleTag{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}
		articleTags := make([]ArticleTag, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			articleTags = append(articleTags, ArticleTag{
				ArticleID: articleID,
				TagID:     tagID,
			})
		}
		return tx.Create(&articleTags).Error
	})
}

// GetTagIDsByArticle 根据文章 ID 获取关联的标签 ID 列表。
func (m *ArticleTagModel) GetTagIDsByArticle(ctx context.Context, articleID string) ([]string, error) {
	var tagIDs []string
	err := m.db.WithContext(ctx).
		Model(&ArticleTag{}).
		Where("article_id = ?", articleID).
		Pluck("tag_id", &tagIDs).Error
	if err != nil {
		return nil, err
	}
	return tagIDs, nil
}
