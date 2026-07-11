package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PortfolioLogic 摄影作品集业务逻辑结构体。
type PortfolioLogic struct {
	portfolioModel     *model.PortfolioModel
	portfolioItemModel *model.PortfolioItemModel
	presetModel        *model.MediaPresetModel
}

// NewPortfolioLogic 创建 PortfolioLogic 实例。
func NewPortfolioLogic() *PortfolioLogic {
	return &PortfolioLogic{
		portfolioModel:     model.NewPortfolio(),
		portfolioItemModel: model.NewPortfolioItem(),
		presetModel:        model.NewMediaPreset(),
	}
}

// Create 创建作品集。
func (l *PortfolioLogic) Create(ctx context.Context, r *req.CreatePortfolioReq) (*res.PortfolioRes, error) {
	portfolio := &model.Portfolio{
		ID:          uuid.New().String(),
		Name:        r.Name,
		Description: r.Description,
	}
	if r.CoverMode != nil {
		portfolio.CoverMode = *r.CoverMode
	}
	if portfolio.CoverMode == 1 {
		if r.CoverPresetID == nil || *r.CoverPresetID == "" {
			return nil, fmt.Errorf("独立封面模式下必须指定封面预设")
		}
		if _, err := l.presetModel.GetByID(ctx, *r.CoverPresetID); err != nil {
			return nil, fmt.Errorf("封面预设不存在")
		}
		portfolio.CoverPresetID = r.CoverPresetID
	}
	if r.Status != nil {
		portfolio.Status = *r.Status
	}
	if r.SortOrder != nil {
		portfolio.SortOrder = *r.SortOrder
	}
	if r.CategoryID != nil {
		portfolio.CategoryID = r.CategoryID
	}

	if err := l.portfolioModel.Create(ctx, portfolio); err != nil {
		return nil, fmt.Errorf("创建作品集失败: %w", err)
	}

	coverMap, _ := l.portfolioModel.ResolveCoverURLs(ctx, []model.Portfolio{*portfolio})
	return l.toPortfolioRes(portfolio, 0, coverMap[portfolio.ID]), nil
}

// GetList 分页查询作品集列表，含每个作品集的作品数量。
func (l *PortfolioLogic) GetList(ctx context.Context, r *req.PortfolioListReq) (*res.PortfolioListRes, error) {
	list, total, err := l.portfolioModel.GetList(ctx, r.Keyword, r.Status, r.CategoryID, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询作品集列表失败: %w", err)
	}

	ids := make([]string, 0, len(list))
	for _, p := range list {
		ids = append(ids, p.ID)
	}
	countMap, err := l.portfolioModel.CountItems(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("统计作品数量失败: %w", err)
	}
	coverMap, err := l.portfolioModel.ResolveCoverURLs(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("解析封面失败: %w", err)
	}

	items := make([]res.PortfolioRes, 0, len(list))
	for _, p := range list {
		items = append(items, *l.toPortfolioRes(&p, countMap[p.ID], coverMap[p.ID]))
	}

	return &res.PortfolioListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询作品集详情（含作品项列表）。
func (l *PortfolioLogic) GetByID(ctx context.Context, id string) (*res.PortfolioDetailRes, error) {
	portfolio, err := l.portfolioModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("作品集不存在")
	}

	items, err := l.portfolioItemModel.GetByPortfolioID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询作品项失败: %w", err)
	}

	itemRes := make([]res.PortfolioItemRes, 0, len(items))
	for _, it := range items {
		itemRes = append(itemRes, *l.toPortfolioItemRes(&it))
	}

	countMap, _ := l.portfolioModel.CountItems(ctx, []string{id})
	coverMap, _ := l.portfolioModel.ResolveCoverURLs(ctx, []model.Portfolio{*portfolio})
	detail := &res.PortfolioDetailRes{
		PortfolioRes: *(l.toPortfolioRes(portfolio, countMap[id], coverMap[portfolio.ID])),
		Items:        itemRes,
	}
	return detail, nil
}

// Update 更新作品集。
func (l *PortfolioLogic) Update(ctx context.Context, id string, r *req.UpdatePortfolioReq) (*res.PortfolioRes, error) {
	portfolio, err := l.portfolioModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("作品集不存在")
	}

	if r.Name != nil {
		portfolio.Name = *r.Name
	}
	if r.Description != nil {
		portfolio.Description = *r.Description
	}
	// 封面设置：cover_mode 与 cover_preset_id 一起处理
	if r.CoverMode != nil || r.CoverPresetID != nil {
		if r.CoverMode != nil {
			portfolio.CoverMode = *r.CoverMode
		}
		if r.CoverPresetID != nil {
			portfolio.CoverPresetID = r.CoverPresetID
		}
		if portfolio.CoverMode == 1 {
			// 独立封面模式：必须有有效预设
			if portfolio.CoverPresetID == nil || *portfolio.CoverPresetID == "" {
				return nil, fmt.Errorf("独立封面模式下必须指定封面预设")
			}
			if _, err := l.presetModel.GetByID(ctx, *portfolio.CoverPresetID); err != nil {
				return nil, fmt.Errorf("封面预设不存在")
			}
		} else {
			// 自动封面模式：清空独立封面预设
			portfolio.CoverPresetID = nil
		}
	}
	if r.Status != nil {
		portfolio.Status = *r.Status
	}
	if r.SortOrder != nil {
		portfolio.SortOrder = *r.SortOrder
	}
	if r.CategoryID != nil {
		portfolio.CategoryID = r.CategoryID
	}

	if err := l.portfolioModel.Update(ctx, portfolio); err != nil {
		return nil, fmt.Errorf("更新作品集失败: %w", err)
	}

	countMap, _ := l.portfolioModel.CountItems(ctx, []string{id})
	coverMap, _ := l.portfolioModel.ResolveCoverURLs(ctx, []model.Portfolio{*portfolio})
	return l.toPortfolioRes(portfolio, countMap[id], coverMap[portfolio.ID]), nil
}

// Delete 删除作品集，事务内级联软删除作品项。
func (l *PortfolioLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.portfolioModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("作品集不存在")
	}

	return l.portfolioItemModel.Transaction(ctx, func(tx *gorm.DB) error {
		// 级联软删除作品项
		if err := tx.Where("portfolio_id = ?", id).Delete(&model.PortfolioItem{}).Error; err != nil {
			return fmt.Errorf("软删除作品项失败: %w", err)
		}
		// 软删除作品集
		if err := tx.Where("id = ?", id).Delete(&model.Portfolio{}).Error; err != nil {
			return fmt.Errorf("软删除作品集失败: %w", err)
		}
		return nil
	})
}

// AddItem 添加作品项，校验作品集与预设存在，sort_order 自增。
func (l *PortfolioLogic) AddItem(ctx context.Context, portfolioID string, r *req.CreatePortfolioItemReq) (*res.PortfolioItemRes, error) {
	// 校验作品集存在
	if _, err := l.portfolioModel.GetByID(ctx, portfolioID); err != nil {
		return nil, fmt.Errorf("作品集不存在")
	}
	// 校验预设存在
	preset, err := l.presetModel.GetByID(ctx, r.PresetID)
	if err != nil {
		return nil, fmt.Errorf("预设不存在")
	}

	maxOrder, err := l.portfolioItemModel.MaxSortOrder(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("查询最大排序失败: %w", err)
	}

	item := &model.PortfolioItem{
		ID:          uuid.New().String(),
		PortfolioID: portfolioID,
		PresetID:    r.PresetID,
		Title:       r.Title,
		Description: r.Description,
		SortOrder:   maxOrder + 1,
	}
	if err := l.portfolioItemModel.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("添加作品项失败: %w", err)
	}

	itemWithPreset := &model.PortfolioItemWithPreset{
		ID:          item.ID,
		PortfolioID: item.PortfolioID,
		PresetID:    item.PresetID,
		Title:       item.Title,
		Description: item.Description,
		SortOrder:   item.SortOrder,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		OutputURL:   preset.OutputURL,
		MimeType:    preset.MimeType,
		OutputSize:  preset.OutputSize,
	}
	return l.toPortfolioItemRes(itemWithPreset), nil
}

// UpdateItem 更新作品项。
func (l *PortfolioLogic) UpdateItem(ctx context.Context, portfolioID, itemID string, r *req.UpdatePortfolioItemReq) (*res.PortfolioItemRes, error) {
	item, err := l.portfolioItemModel.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("作品项不存在")
	}
	if item.PortfolioID != portfolioID {
		return nil, fmt.Errorf("作品项不存在")
	}

	if r.PresetID != nil {
		// 校验预设存在
		if _, err := l.presetModel.GetByID(ctx, *r.PresetID); err != nil {
			return nil, fmt.Errorf("预设不存在")
		}
		item.PresetID = *r.PresetID
	}
	if r.Title != nil {
		item.Title = *r.Title
	}
	if r.Description != nil {
		item.Description = *r.Description
	}

	if err := l.portfolioItemModel.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("更新作品项失败: %w", err)
	}

	items, err := l.portfolioItemModel.GetByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("查询作品项失败: %w", err)
	}
	for i := range items {
		if items[i].ID == itemID {
			return l.toPortfolioItemRes(&items[i]), nil
		}
	}
	return l.toPortfolioItemRes(&model.PortfolioItemWithPreset{
		ID:          item.ID,
		PortfolioID: item.PortfolioID,
		PresetID:    item.PresetID,
		Title:       item.Title,
		Description: item.Description,
		SortOrder:   item.SortOrder,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}), nil
}

// DeleteItem 删除作品项。
func (l *PortfolioLogic) DeleteItem(ctx context.Context, portfolioID, itemID string) error {
	item, err := l.portfolioItemModel.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("作品项不存在")
	}
	if item.PortfolioID != portfolioID {
		return fmt.Errorf("作品项不存在")
	}
	return l.portfolioItemModel.SoftDelete(ctx, itemID)
}

// SortItems 批量更新作品项排序，事务内执行。
func (l *PortfolioLogic) SortItems(ctx context.Context, portfolioID string, r *req.SortPortfolioItemsReq) error {
	if _, err := l.portfolioModel.GetByID(ctx, portfolioID); err != nil {
		return fmt.Errorf("作品集不存在")
	}

	items := make([]model.PortfolioItem, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, model.PortfolioItem{
			ID:        it.ID,
			SortOrder: it.SortOrder,
		})
	}

	return l.portfolioItemModel.Transaction(ctx, func(tx *gorm.DB) error {
		for _, it := range items {
			if err := tx.Model(&model.PortfolioItem{}).
				Where("id = ? AND portfolio_id = ?", it.ID, portfolioID).
				Update("sort_order", it.SortOrder).Error; err != nil {
				return fmt.Errorf("更新排序失败: %w", err)
			}
		}
		return nil
	})
}

// toPortfolioRes 转换为作品集响应。
func (l *PortfolioLogic) toPortfolioRes(p *model.Portfolio, count int64, coverURL string) *res.PortfolioRes {
	coverID := ""
	if p.CoverPresetID != nil {
		coverID = *p.CoverPresetID
	}
	categoryID := ""
	categoryName := ""
	if p.CategoryID != nil {
		categoryID = *p.CategoryID
	}
	if p.Category != nil {
		categoryName = p.Category.Name
	}
	return &res.PortfolioRes{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		CoverMode:     p.CoverMode,
		CoverPresetID: coverID,
		CoverURL:      coverURL,
		Status:        p.Status,
		SortOrder:     p.SortOrder,
		CategoryID:    categoryID,
		CategoryName:  categoryName,
		ItemCount:     count,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// toPortfolioItemRes 转换为作品项响应。
func (l *PortfolioLogic) toPortfolioItemRes(it *model.PortfolioItemWithPreset) *res.PortfolioItemRes {
	return &res.PortfolioItemRes{
		ID:          it.ID,
		PortfolioID: it.PortfolioID,
		PresetID:    it.PresetID,
		Title:       it.Title,
		Description: it.Description,
		SortOrder:   it.SortOrder,
		OutputURL:   it.OutputURL,
		MimeType:    it.MimeType,
		OutputSize:  it.OutputSize,
		CreatedAt:   it.CreatedAt,
		UpdatedAt:   it.UpdatedAt,
	}
}
