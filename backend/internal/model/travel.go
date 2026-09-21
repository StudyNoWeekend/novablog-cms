package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TravelGuide 旅行攻略模型，对应 travel_guides 表。
type TravelGuide struct {
	ID          string           `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string           `gorm:"type:varchar(200);not null" json:"title"`
	Summary     string           `gorm:"type:varchar(500)" json:"summary"`
	CoverImage  string           `gorm:"type:varchar(500);column:cover_image" json:"cover_image"`
	Status      int16            `gorm:"type:smallint;default:1" json:"status"` // 1=草稿 2=已发布 3=已下架
	Destination string           `gorm:"type:varchar(200);not null" json:"destination"`
	Region      string           `gorm:"type:varchar(50);not null" json:"region"`
	CategoryID  *string          `gorm:"type:uuid;column:category_id" json:"category_id"`
	Category    *Category        `gorm:"foreignKey:CategoryID" json:"-"`
	Days        int              `gorm:"type:int;default:1" json:"days"`
	BestMonth   string           `gorm:"type:varchar(100);column:best_month" json:"best_month"`
	ViewCount   int              `gorm:"type:int;default:0;column:view_count" json:"view_count"`
	LikeCount   int              `gorm:"type:int;default:0;column:like_count" json:"like_count"`
	Rating      float64          `gorm:"type:decimal(2,1);default:0.0" json:"rating"`
	ReviewCount int              `gorm:"type:int;default:0;column:review_count" json:"review_count"`
	Attractions []map[string]any `gorm:"type:jsonb;serializer:json" json:"attractions"`
	Itinerary   []map[string]any `gorm:"type:jsonb;serializer:json" json:"itinerary"`
	Reviews     []map[string]any `gorm:"type:jsonb;serializer:json" json:"reviews"`
	CreatedAt   time.Time        `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定数据表名称。
func (TravelGuide) TableName() string {
	return "travel_guides"
}

// AfterFind GORM 查询后钩子，将相对路径 URL 拼接为完整 URL。
func (t *TravelGuide) AfterFind(tx *gorm.DB) error {
	t.CoverImage = resolveURL(t.CoverImage)
	return nil
}

// TravelGuideModel 旅行攻略模型操作结构体。
type TravelGuideModel struct {
	db *gorm.DB
}

// NewTravelGuide 创建 TravelGuideModel 实例。
func NewTravelGuide() *TravelGuideModel {
	return &TravelGuideModel{db: DB}
}

// Create 创建旅行攻略。
func (m *TravelGuideModel) Create(ctx context.Context, guide *TravelGuide) error {
	return m.db.WithContext(ctx).Create(guide).Error
}

// GetByID 根据 ID 查询旅行攻略。
func (m *TravelGuideModel) GetByID(ctx context.Context, id string) (*TravelGuide, error) {
	var guide TravelGuide
	err := m.db.WithContext(ctx).Preload("Category").Where("id = ?", id).First(&guide).Error
	if err != nil {
		return nil, err
	}
	return &guide, nil
}

// GetList 分页查询旅行攻略列表，支持 keyword 模糊搜索、region/status 精确匹配、daysRange 筛选和多种排序。
func (m *TravelGuideModel) GetList(ctx context.Context, page, pageSize int, keyword, region *string, status *int16, daysRange *string, categoryID *string, sort string) ([]TravelGuide, int64, error) {
	query := m.db.WithContext(ctx).Model(&TravelGuide{})

	if keyword != nil && *keyword != "" {
		like := "%" + *keyword + "%"
		query = query.Where("title ILIKE ? OR destination ILIKE ? OR summary ILIKE ?", like, like, like)
	}
	if region != nil && *region != "" {
		query = query.Where("region = ?", *region)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if daysRange != nil && *daysRange != "" && *daysRange != "all" {
		switch *daysRange {
		case "1-3":
			query = query.Where("days >= 1 AND days <= 3")
		case "4-7":
			query = query.Where("days >= 4 AND days <= 7")
		case "8-14":
			query = query.Where("days >= 8 AND days <= 14")
		case "15+":
			query = query.Where("days >= 15")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC"
	switch sort {
	case "views":
		orderClause = "view_count DESC"
	case "rating":
		orderClause = "rating DESC"
	case "likes":
		orderClause = "like_count DESC"
	default:
		orderClause = "created_at DESC"
	}

	offset := (page - 1) * pageSize
	var guides []TravelGuide
	err := query.
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&guides).Error
	if err != nil {
		return nil, 0, err
	}
	return guides, total, nil
}

// Update 更新旅行攻略。
func (m *TravelGuideModel) Update(ctx context.Context, guide *TravelGuide) error {
	return m.db.WithContext(ctx).Save(guide).Error
}

// UpdateStatus 更新旅行攻略状态。
func (m *TravelGuideModel) UpdateStatus(ctx context.Context, id string, status int16) error {
	return m.db.WithContext(ctx).Model(&TravelGuide{}).Where("id = ?", id).Update("status", status).Error
}

// SoftDelete 软删除旅行攻略。
func (m *TravelGuideModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&TravelGuide{}).Error
}

// GetHotList 查询热门旅行攻略列表，按浏览量降序取指定数量。
func (m *TravelGuideModel) GetHotList(ctx context.Context, count int) ([]TravelGuide, error) {
	var guides []TravelGuide
	err := m.db.WithContext(ctx).
		Where("status = ?", 2).
		Order("view_count DESC").
		Limit(count).
		Find(&guides).Error
	if err != nil {
		return nil, err
	}
	return guides, nil
}

// IncrementViewCount 根据ID增加旅行攻略浏览量，仅对已发布攻略生效。
func (m *TravelGuideModel) IncrementViewCount(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).
		Model(&TravelGuide{}).
		Where("id = ? AND status = ?", id, 2).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateReviewCount 增减旅行攻略评论计数，仅对已发布攻略生效。delta 为 +1（新增）或 -1（删除）。
func (m *TravelGuideModel) UpdateReviewCount(ctx context.Context, id string, delta int) error {
	expr := gorm.Expr("review_count + ?", delta)
	if delta < 0 {
		expr = gorm.Expr("GREATEST(review_count - ?, 0)", -delta)
	}
	return m.db.WithContext(ctx).
		Model(&TravelGuide{}).
		Where("id = ? AND status = ?", id, 2).
		UpdateColumn("review_count", expr).Error
}

// IncrementLikeCount 根据ID增加旅行攻略点赞数，仅对已发布攻略生效。
func (m *TravelGuideModel) IncrementLikeCount(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).
		Model(&TravelGuide{}).
		Where("id = ? AND status = ?", id, 2).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}
