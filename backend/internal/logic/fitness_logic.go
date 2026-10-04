package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// FitnessLogic 健身训练业务逻辑结构体。
type FitnessLogic struct {
	recordModel *model.FitnessRecordModel
}

// NewFitnessLogic 创建 FitnessLogic 实例。
func NewFitnessLogic() *FitnessLogic {
	return &FitnessLogic{recordModel: model.NewFitnessRecord()}
}

// Create 创建训练记录。
func (l *FitnessLogic) Create(ctx context.Context, r *req.CreateFitnessReq) (*res.FitnessRes, error) {
	date := time.Now()
	if r.Date != nil {
		date = *r.Date
	}
	m := &model.FitnessRecord{
		ID:      uuid.New().String(),
		Date:    date,
		Title:   strings.TrimSpace(r.Title),
		Type:    "strength",
		Content: emptyIfNil(r.Content),
		Notes:   r.Notes,
	}
	if r.Type != "" {
		m.Type = r.Type
	}
	if r.DurationMin != nil {
		m.DurationMin = *r.DurationMin
	}
	if r.Calories != nil {
		m.Calories = *r.Calories
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.recordModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建训练记录失败: %w", err)
	}
	return l.toRes(m), nil
}

// GetList 分页查询训练记录列表。
func (l *FitnessLogic) GetList(ctx context.Context, r *req.FitnessListReq) (*res.FitnessListRes, error) {
	list, total, err := l.recordModel.GetList(ctx, r.Keyword, r.Type, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询训练记录列表失败: %w", err)
	}

	items := make([]res.FitnessCardRes, 0, len(list))
	for i := range list {
		items = append(items, l.toCard(&list[i]))
	}
	return &res.FitnessListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询训练记录详情。
func (l *FitnessLogic) GetByID(ctx context.Context, id string) (*res.FitnessRes, error) {
	m, err := l.recordModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("训练记录不存在")
	}
	return l.toRes(m), nil
}

// Update 更新训练记录（部分更新语义）。
func (l *FitnessLogic) Update(ctx context.Context, id string, r *req.UpdateFitnessReq) (*res.FitnessRes, error) {
	m, err := l.recordModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("训练记录不存在")
	}

	if r.Date != nil {
		m.Date = *r.Date
	}
	if r.Title != nil {
		m.Title = strings.TrimSpace(*r.Title)
	}
	if r.Type != nil {
		m.Type = *r.Type
	}
	if r.DurationMin != nil {
		m.DurationMin = *r.DurationMin
	}
	if r.Calories != nil {
		m.Calories = *r.Calories
	}
	if r.Content != nil {
		m.Content = r.Content
	}
	if r.Notes != nil {
		m.Notes = *r.Notes
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.recordModel.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新训练记录失败: %w", err)
	}
	return l.toRes(m), nil
}

// Delete 删除训练记录。
func (l *FitnessLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.recordModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("训练记录不存在")
	}
	if err := l.recordModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除训练记录失败: %w", err)
	}
	return nil
}

// GetPublicList 获取已发布训练记录列表（强制 status=1）。
func (l *FitnessLogic) GetPublicList(ctx context.Context, r *req.FitnessListReq) (*res.FitnessListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// toRes 转换为训练记录详情响应。
func (l *FitnessLogic) toRes(m *model.FitnessRecord) *res.FitnessRes {
	return &res.FitnessRes{
		ID:          m.ID,
		Date:        m.Date,
		Title:       m.Title,
		Type:        m.Type,
		DurationMin: m.DurationMin,
		Calories:    m.Calories,
		Content:     m.Content,
		Notes:       m.Notes,
		Status:      m.Status,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toCard 转换为训练记录列表卡片响应。
func (l *FitnessLogic) toCard(m *model.FitnessRecord) res.FitnessCardRes {
	return res.FitnessCardRes{
		ID:          m.ID,
		Date:        m.Date,
		Title:       m.Title,
		Type:        m.Type,
		DurationMin: m.DurationMin,
		Calories:    m.Calories,
		Status:      m.Status,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
