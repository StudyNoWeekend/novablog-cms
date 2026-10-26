package logic

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// TechStackLogic 技术栈业务逻辑结构体。
type TechStackLogic struct {
	itemModel     *model.TechStackItemModel
	categoryModel *model.CategoryModel
}

// NewTechStackLogic 创建 TechStackLogic 实例。
func NewTechStackLogic() *TechStackLogic {
	return &TechStackLogic{
		itemModel:     model.NewTechStackItem(),
		categoryModel: model.NewCategory(),
	}
}

// Create 创建技术栈条目。
func (l *TechStackLogic) Create(ctx context.Context, r *req.CreateTechStackReq) (*res.TechStackItemRes, error) {
	iconMediaID, _, icon, err := normalizeMediaRef(ctx, r.IconMediaID, "", r.Icon)
	if err != nil {
		return nil, err
	}
	m := &model.TechStackItem{
		ID:          uuid.New().String(),
		Name:        strings.TrimSpace(r.Name),
		Category:    r.Category,
		Icon:        icon,
		IconMediaID: mediaRefPtr(iconMediaID),
		Description: r.Description,
	}
	if r.Level != nil {
		m.Level = *r.Level
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}
	if r.CategoryID != nil && *r.CategoryID != "" {
		// 结构化分类优先：legacy Category 镜像为分类 slug，保证公开 ?category= 筛选与主题分组兼容
		cat, fetchErr := l.categoryModel.GetByID(ctx, *r.CategoryID)
		if fetchErr != nil {
			return nil, fmt.Errorf("分类不存在")
		}
		m.CategoryID = r.CategoryID
		m.Category = cat.Slug
	}

	if err := l.itemModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建技术栈条目失败: %w", err)
	}
	result := l.toRes(m)
	return &result, nil
}

// GetList 分页查询技术栈条目列表。
func (l *TechStackLogic) GetList(ctx context.Context, r *req.TechStackListReq) (*res.TechStackListRes, error) {
	list, total, err := l.itemModel.GetList(ctx, r.Keyword, r.Category, r.CategoryID, r.Level, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询技术栈列表失败: %w", err)
	}

	items := make([]res.TechStackItemRes, 0, len(list))
	l.resolveIconURLs(ctx, list)
	for i := range list {
		items = append(items, l.toRes(&list[i]))
	}
	return &res.TechStackListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询技术栈条目详情。
func (l *TechStackLogic) GetByID(ctx context.Context, id string) (*res.TechStackItemRes, error) {
	m, err := l.itemModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("技术栈条目不存在")
	}
	l.resolveIconURL(ctx, m)
	result := l.toRes(m)
	return &result, nil
}

// Update 更新技术栈条目（部分更新语义）。
func (l *TechStackLogic) Update(ctx context.Context, id string, r *req.UpdateTechStackReq) (*res.TechStackItemRes, error) {
	m, err := l.itemModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("技术栈条目不存在")
	}

	if r.Name != nil {
		m.Name = strings.TrimSpace(*r.Name)
	}
	if r.Category != nil {
		m.Category = *r.Category
	}
	if r.CategoryID != nil && *r.CategoryID != "" {
		// 结构化分类优先：legacy Category 镜像为分类 slug
		cat, fetchErr := l.categoryModel.GetByID(ctx, *r.CategoryID)
		if fetchErr != nil {
			return nil, fmt.Errorf("分类不存在")
		}
		m.CategoryID = r.CategoryID
		m.Category = cat.Slug
	}
	if r.Icon != nil || r.IconMediaID != nil {
		mediaID := ""
		if r.IconMediaID != nil {
			mediaID = *r.IconMediaID
		}
		iconURL := ""
		if r.Icon != nil {
			iconURL = *r.Icon
		}
		var iconMediaID, icon string
		iconMediaID, _, icon, err = normalizeMediaRef(ctx, mediaID, "", iconURL)
		if err != nil {
			return nil, err
		}
		m.Icon = icon
		m.IconMediaID = mediaRefPtr(iconMediaID)
	}
	if r.Level != nil {
		m.Level = *r.Level
	}
	if r.Description != nil {
		m.Description = *r.Description
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.itemModel.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新技术栈条目失败: %w", err)
	}
	result := l.toRes(m)
	return &result, nil
}

// Delete 删除技术栈条目。
func (l *TechStackLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.itemModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("技术栈条目不存在")
	}
	if err := l.itemModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除技术栈条目失败: %w", err)
	}
	return nil
}

// GetPublicList 获取已发布技术栈条目列表（强制 status=1）。
func (l *TechStackLogic) GetPublicList(ctx context.Context, r *req.TechStackListReq) (*res.TechStackListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// resolveIconURLs 批量解析技术栈图标媒体引用为当前完整 URL（外链行保持原值）。
func (l *TechStackLogic) resolveIconURLs(ctx context.Context, items []model.TechStackItem) {
	ptrs := make([]*model.TechStackItem, len(items))
	for i := range items {
		ptrs[i] = &items[i]
	}
	model.ResolveCoverURLs(ctx, ptrs,
		func(t *model.TechStackItem) (string, string) { return refStr(t.IconMediaID), "" },
		func(t *model.TechStackItem, u string) { t.Icon = u },
	)
}

// resolveIconURL 解析单条技术栈图标媒体引用。
func (l *TechStackLogic) resolveIconURL(ctx context.Context, t *model.TechStackItem) {
	model.ResolveCoverURLs(ctx, []*model.TechStackItem{t},
		func(x *model.TechStackItem) (string, string) { return refStr(x.IconMediaID), "" },
		func(x *model.TechStackItem, u string) { x.Icon = u },
	)
}

// toRes 转换为技术栈条目响应。
func (l *TechStackLogic) toRes(m *model.TechStackItem) res.TechStackItemRes {
	return res.TechStackItemRes{
		ID:           m.ID,
		Name:         m.Name,
		Category:     m.Category,
		CategoryID:   categoryIDOf(m.CategoryID),
		CategoryName: categoryNameOf(m.CategoryInfo),
		Icon:         m.Icon,
		IconMediaID:  refStr(m.IconMediaID),
		Level:        m.Level,
		Description:  m.Description,
		Status:       m.Status,
		SortOrder:    m.SortOrder,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
