package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Book 读书书架模型，对应 books 数据表。
type Book struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Author        string         `gorm:"type:varchar(255);default:''" json:"author"`
	Cover         string         `gorm:"type:varchar(500)" json:"cover"`
	Rating        int            `gorm:"type:int;default:0" json:"rating"`                                            // 0-5 星，0=未评分
	ReadingStatus string         `gorm:"column:reading_status;type:varchar(20);default:'want'" json:"reading_status"` // want=想读, reading=在读, done=读完
	Review        string         `gorm:"type:text" json:"review"`                                                     // 书评（Markdown）
	StartedAt     *time.Time     `gorm:"type:timestamptz" json:"started_at"`
	FinishedAt    *time.Time     `gorm:"type:timestamptz" json:"finished_at"`
	Status        int            `gorm:"type:int;default:0" json:"status"` // 0=草稿, 1=已发布
	SortOrder     int            `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	CreatedAt     time.Time      `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据表名称。
func (Book) TableName() string {
	return "books"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (b *Book) AfterFind(tx *gorm.DB) error {
	b.Cover = resolveURL(b.Cover)
	return nil
}

// BookModel 读书书架模型操作结构体。
type BookModel struct {
	db *gorm.DB
}

// NewBook 创建 BookModel 实例。
func NewBook() *BookModel {
	return &BookModel{db: DB}
}

// Create 创建书籍记录。
func (m *BookModel) Create(ctx context.Context, b *Book) error {
	return m.db.WithContext(ctx).Create(b).Error
}

// GetByID 根据 ID 查询书籍。
func (m *BookModel) GetByID(ctx context.Context, id string) (*Book, error) {
	var b Book
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetList 分页查询书籍列表，支持 keyword 模糊搜索书名/作者、readingStatus 筛选、status 过滤。
func (m *BookModel) GetList(ctx context.Context, keyword *string, readingStatus *string, status *int, page, pageSize int) ([]Book, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Book{})

	if keyword != nil && *keyword != "" {
		like := "%" + *keyword + "%"
		query = query.Where("title ILIKE ? OR author ILIKE ?", like, like)
	}
	if readingStatus != nil && *readingStatus != "" && *readingStatus != "all" {
		query = query.Where("reading_status = ?", *readingStatus)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Book
	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新书籍。
func (m *BookModel) Update(ctx context.Context, b *Book) error {
	return m.db.WithContext(ctx).Save(b).Error
}

// SoftDelete 软删除书籍。
func (m *BookModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Book{}).Error
}
