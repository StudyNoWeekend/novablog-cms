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
	itemModel *model.TechStackItemModel
}

// NewTechStackLogic 创建 TechStackLogic 实例。
func NewTechStackLogic() *TechStackLogic {
	return &TechStackLogic{itemModel: model.NewTechStackItem()}
}

// Create 创建技术栈条目。
func (l *TechStackLogic) Create(ctx context.Context, r *req.CreateTechStackReq) (*res.TechStackItemRes, error) {
	m := &model.TechStackItem{
		ID:          uuid.New().String(),
		Name:        strings.TrimSpace(r.Name),
		Category:    r.Category,
		Icon:        r.Icon,
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

	if err := l.itemModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建技术栈条目失败: %w", err)
	}
	result := l.toRes(m)
	return &result, nil
}

// GetList 分页查询技术栈条目列表。
func (l *TechStackLogic) GetList(ctx context.Context, r *req.TechStackListReq) (*res.TechStackListRes, error) {
	list, total, err := l.itemModel.GetList(ctx, r.Keyword, r.Category, r.Level, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询技术栈列表失败: %w", err)
	}

	items := make([]res.TechStackItemRes, 0, len(list))
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
	if r.Icon != nil {
		m.Icon = *r.Icon
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

// toRes 转换为技术栈条目响应。
func (l *TechStackLogic) toRes(m *model.TechStackItem) res.TechStackItemRes {
	return res.TechStackItemRes{
		ID:          m.ID,
		Name:        m.Name,
		Category:    m.Category,
		Icon:        m.Icon,
		Level:       m.Level,
		Description: m.Description,
		Status:      m.Status,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
