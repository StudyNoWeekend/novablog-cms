package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Portfolio 摄影作品集模型，对应 portfolios 数据表。
type Portfolio struct {
	ID            string         `gorm:"type:uuid;primaryKey"`
	Name          string         `gorm:"type:varchar(255);not null"`
	Description   string         `gorm:"type:text"`
	CoverMode     int            `gorm:"column:cover_mode;type:int;default:0"` // 0=使用排序第一的作品, 1=独立设置封面
	CoverPresetID *string        `gorm:"column:cover_preset_id;type:uuid"`
	Status        int            `gorm:"type:int;default:0"` // 0=草稿, 1=已发布
	SortOrder     int            `gorm:"column:sort_order;type:int;default:0"`
	CategoryID    *string        `gorm:"type:uuid;column:category_id" json:"category_id"`
	Category      *Category      `gorm:"foreignKey:CategoryID" json:"-"`
	CreatedAt     time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (Portfolio) TableName() string {
	return "portfolios"
}

// PortfolioModel 作品集模型操作结构体。
type PortfolioModel struct {
	db *gorm.DB
}

// NewPortfolio 创建 PortfolioModel 实例。
func NewPortfolio() *PortfolioModel {
	return &PortfolioModel{db: DB}
}

// Create 创建作品集记录。
func (m *PortfolioModel) Create(ctx context.Context, portfolio *Portfolio) error {
	return m.db.WithContext(ctx).Create(portfolio).Error
}

// GetByID 根据 ID 查询作品集。
func (m *PortfolioModel) GetByID(ctx context.Context, id string) (*Portfolio, error) {
	var portfolio Portfolio
	err := m.db.WithContext(ctx).Preload("Category").Where("id = ?", id).First(&portfolio).Error
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

// GetList 分页查询作品集列表，支持按 keyword 模糊搜索名称、按 status 过滤。
func (m *PortfolioModel) GetList(ctx context.Context, keyword *string, status *int, categoryID *string, page, pageSize int) ([]Portfolio, int64, error) {
	var total int64
	query := m.db.WithContext(ctx).Model(&Portfolio{})

	if keyword != nil && *keyword != "" {
		query = query.Where("name ILIKE ?", "%"+*keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Portfolio
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

// Update 更新作品集。
func (m *PortfolioModel) Update(ctx context.Context, portfolio *Portfolio) error {
	return m.db.WithContext(ctx).Save(portfolio).Error
}

// SoftDelete 软删除作品集。
func (m *PortfolioModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&Portfolio{}).Error
}

// CountItems 批量统计多个作品集下的作品数量，返回 map[portfolioID]count。
func (m *PortfolioModel) CountItems(ctx context.Context, portfolioIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(portfolioIDs) == 0 {
		return result, nil
	}

	type row struct {
		PortfolioID string `gorm:"column:portfolio_id"`
		Count       int64  `gorm:"column:count"`
	}
	var rows []row
	err := m.db.WithContext(ctx).
		Table("portfolio_items").
		Select("portfolio_id, COUNT(*) AS count").
		Where("portfolio_id IN ? AND deleted_at IS NULL", portfolioIDs).
		Group("portfolio_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.PortfolioID] = r.Count
	}
	return result, nil
}

// ResolveCoverURLs 批量解析作品集封面 URL，返回 map[portfolioID]coverURL。
// cover_mode=0：取排序第一的作品项关联预设的 output_url；
// cover_mode=1：取 cover_preset_id 对应预设的 output_url。
func (m *PortfolioModel) ResolveCoverURLs(ctx context.Context, portfolios []Portfolio) (map[string]string, error) {
	result := make(map[string]string)
	if len(portfolios) == 0 {
		return result, nil
	}

	autoIDs := make([]string, 0)
	customPresetIDs := make([]string, 0)
	for _, p := range portfolios {
		if p.CoverMode == 1 && p.CoverPresetID != nil && *p.CoverPresetID != "" {
			customPresetIDs = append(customPresetIDs, *p.CoverPresetID)
		} else {
			autoIDs = append(autoIDs, p.ID)
		}
	}

	// 独立封面：批量查询 media_presets 取 output_url
	if len(customPresetIDs) > 0 {
		type presetRow struct {
			ID        string `gorm:"column:id"`
			OutputURL string `gorm:"column:output_url"`
		}
		var presetRows []presetRow
		err := m.db.WithContext(ctx).
			Table("media_presets").
			Select("id, output_url").
			Where("id IN ? AND deleted_at IS NULL", customPresetIDs).
			Scan(&presetRows).Error
		if err != nil {
			return nil, err
		}
		presetURLMap := make(map[string]string, len(presetRows))
		for _, r := range presetRows {
			presetURLMap[r.ID] = r.OutputURL
		}
		for _, p := range portfolios {
			if p.CoverMode == 1 && p.CoverPresetID != nil {
				if url, ok := presetURLMap[*p.CoverPresetID]; ok {
					result[p.ID] = url
				}
			}
		}
	}

	// 自动封面：取每个作品集排序第一的作品项预设 output_url
	if len(autoIDs) > 0 {
		type itemRow struct {
			PortfolioID string `gorm:"column:portfolio_id"`
			OutputURL   string `gorm:"column:output_url"`
		}
		var itemRows []itemRow
		err := m.db.WithContext(ctx).
			Table("portfolio_items pi").
			Select("DISTINCT ON (pi.portfolio_id) pi.portfolio_id, mp.output_url").
			Joins("LEFT JOIN media_presets mp ON mp.id = pi.preset_id AND mp.deleted_at IS NULL").
			Where("pi.portfolio_id IN ? AND pi.deleted_at IS NULL", autoIDs).
			Order("pi.portfolio_id, pi.sort_order ASC, pi.created_at ASC").
			Scan(&itemRows).Error
		if err != nil {
			return nil, err
		}
		for _, r := range itemRows {
			result[r.PortfolioID] = r.OutputURL
		}
	}

	return result, nil
}

// PortfolioItem 作品项模型，对应 portfolio_items 数据表。
type PortfolioItem struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	PortfolioID string         `gorm:"column:portfolio_id;type:uuid;not null"`
	PresetID    string         `gorm:"column:preset_id;type:uuid;not null"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	SortOrder   int            `gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (PortfolioItem) TableName() string {
	return "portfolio_items"
}

// PortfolioItemWithPreset 作品项关联预设信息的查询结果结构体。
type PortfolioItemWithPreset struct {
	ID          string    `gorm:"column:id"`
	PortfolioID string    `gorm:"column:portfolio_id"`
	PresetID    string    `gorm:"column:preset_id"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	SortOrder   int       `gorm:"column:sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	OutputURL   string    `gorm:"column:output_url"`
	MimeType    string    `gorm:"column:mime_type"`
	OutputSize  int64     `gorm:"column:output_size"`
}

// PortfolioItemModel 作品项模型操作结构体。
type PortfolioItemModel struct {
	db *gorm.DB
}

// NewPortfolioItem 创建 PortfolioItemModel 实例。
func NewPortfolioItem() *PortfolioItemModel {
	return &PortfolioItemModel{db: DB}
}

// Create 创建作品项记录。
func (m *PortfolioItemModel) Create(ctx context.Context, item *PortfolioItem) error {
	return m.db.WithContext(ctx).Create(item).Error
}

// GetByPortfolioID 查询指定作品集下的所有作品项，LEFT JOIN media_presets 取出成品图信息，按 sort_order 升序。
func (m *PortfolioItemModel) GetByPortfolioID(ctx context.Context, portfolioID string) ([]PortfolioItemWithPreset, error) {
	var items []PortfolioItemWithPreset
	err := m.db.WithContext(ctx).
		Table("portfolio_items pi").
		Select("pi.id, pi.portfolio_id, pi.preset_id, pi.title, pi.description, pi.sort_order, pi.created_at, pi.updated_at, mp.output_url, mp.mime_type, mp.output_size").
		Joins("LEFT JOIN media_presets mp ON mp.id = pi.preset_id AND mp.deleted_at IS NULL").
		Where("pi.portfolio_id = ? AND pi.deleted_at IS NULL", portfolioID).
		Order("pi.sort_order ASC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetByID 根据 ID 查询作品项。
func (m *PortfolioItemModel) GetByID(ctx context.Context, id string) (*PortfolioItem, error) {
	var item PortfolioItem
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update 更新作品项。
func (m *PortfolioItemModel) Update(ctx context.Context, item *PortfolioItem) error {
	return m.db.WithContext(ctx).Save(item).Error
}

// SoftDelete 软删除作品项。
func (m *PortfolioItemModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&PortfolioItem{}).Error
}

// SoftDeleteByPortfolioID 按作品集 ID 级联软删除所有作品项。
func (m *PortfolioItemModel) SoftDeleteByPortfolioID(ctx context.Context, portfolioID string) error {
	return m.db.WithContext(ctx).
		Where("portfolio_id = ?", portfolioID).
		Delete(&PortfolioItem{}).Error
}

// MaxSortOrder 返回指定作品集下最大 sort_order，若无则为 0。
func (m *PortfolioItemModel) MaxSortOrder(ctx context.Context, portfolioID string) (int, error) {
	var maxOrder int
	err := m.db.WithContext(ctx).
		Model(&PortfolioItem{}).
		Where("portfolio_id = ?", portfolioID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}

// BatchUpdateSort 在事务内批量更新作品项的 sort_order。
func (m *PortfolioItemModel) BatchUpdateSort(ctx context.Context, tx *gorm.DB, items []PortfolioItem) error {
	for _, item := range items {
		if err := tx.Model(&PortfolioItem{}).
			Where("id = ?", item.ID).
			Update("sort_order", item.SortOrder).Error; err != nil {
			return err
		}
	}
	return nil
}

// Transaction 暴露事务能力，供 logic 层使用。
func (m *PortfolioItemModel) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}
