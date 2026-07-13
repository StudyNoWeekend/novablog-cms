package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// TravelGuideLogic 旅行攻略业务逻辑结构体。
type TravelGuideLogic struct {
	model *model.TravelGuideModel
}

// NewTravelGuideLogic 创建 TravelGuideLogic 实例。
func NewTravelGuideLogic() *TravelGuideLogic {
	return &TravelGuideLogic{
		model: model.NewTravelGuide(),
	}
}

// Create 创建旅行攻略，生成 UUID 并设置默认值后创建记录。
func (l *TravelGuideLogic) Create(ctx context.Context, r *req.CreateTravelGuideReq) (*res.TravelGuideDetailRes, error) {
	status := r.Status
	if status == 0 {
		status = 1 // 默认草稿
	}
	days := r.Days
	if days <= 0 {
		days = 1
	}

	guide := &model.TravelGuide{
		ID:          uuid.New().String(),
		Title:       r.Title,
		Summary:     r.Summary,
		CoverImage:  r.CoverImage,
		Status:      status,
		Destination: r.Destination,
		Region:      r.Region,
		CategoryID:  r.CategoryID,
		Days:        days,
		BestMonth:   r.BestMonth,
		ViewCount:   0,
		LikeCount:   0,
		Rating:      0,
		ReviewCount: 0,
		Attractions: r.Attractions,
		Itinerary:   r.Itinerary,
		Reviews:     r.Reviews,
	}

	if err := l.model.Create(ctx, guide); err != nil {
		return nil, fmt.Errorf("创建旅行攻略失败: %w", err)
	}

	return l.GetDetail(ctx, guide.ID)
}

// Update 更新旅行攻略，按指针字段逐项更新。
func (l *TravelGuideLogic) Update(ctx context.Context, id string, r *req.UpdateTravelGuideReq) (*res.TravelGuideDetailRes, error) {
	guide, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("旅行攻略不存在")
	}

	if r.Title != nil {
		guide.Title = *r.Title
	}
	if r.Summary != nil {
		guide.Summary = *r.Summary
	}
	if r.CoverImage != nil {
		guide.CoverImage = *r.CoverImage
	}
	if r.Status != nil {
		guide.Status = *r.Status
	}
	if r.Destination != nil {
		guide.Destination = *r.Destination
	}
	if r.Region != nil {
		guide.Region = *r.Region
	}
	if r.CategoryID != nil {
		guide.CategoryID = r.CategoryID
	}
	if r.Days != nil {
		guide.Days = *r.Days
	}
	if r.BestMonth != nil {
		guide.BestMonth = *r.BestMonth
	}
	if r.Attractions != nil {
		guide.Attractions = r.Attractions
	}
	if r.Itinerary != nil {
		guide.Itinerary = r.Itinerary
	}
	if r.Reviews != nil {
		guide.Reviews = r.Reviews
	}

	if err := l.model.Update(ctx, guide); err != nil {
		return nil, fmt.Errorf("更新旅行攻略失败: %w", err)
	}

	return l.GetDetail(ctx, id)
}

// GetList 获取旅行攻略列表，支持分页、搜索、筛选和排序。
func (l *TravelGuideLogic) GetList(ctx context.Context, r *req.TravelGuideListReq) (*res.PageRes[res.TravelGuideRes], error) {
	sort := "latest"
	if r.Sort != nil && *r.Sort != "" {
		sort = *r.Sort
	}

	guides, total, err := l.model.GetList(
		ctx, r.GetPage(), r.GetPageSize(),
		r.Keyword, r.Region, r.Status, r.DaysRange, r.CategoryID, sort,
	)
	if err != nil {
		return nil, err
	}

	var items []res.TravelGuideRes
	for i := range guides {
		items = append(items, l.toGuideRes(&guides[i]))
	}

	return res.NewPageRes(items, total, r.GetPage(), r.GetPageSize()), nil
}

// GetDetail 获取旅行攻略详情。
func (l *TravelGuideLogic) GetDetail(ctx context.Context, id string) (*res.TravelGuideDetailRes, error) {
	guide, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("旅行攻略不存在")
	}
	return l.toGuideDetailRes(guide), nil
}

// UpdateStatus 更新旅行攻略状态。
func (l *TravelGuideLogic) UpdateStatus(ctx context.Context, id string, r *req.UpdateTravelGuideStatusReq) error {
	_, err := l.model.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("旅行攻略不存在")
	}
	return l.model.UpdateStatus(ctx, id, r.Status)
}

// Delete 软删除旅行攻略。
func (l *TravelGuideLogic) Delete(ctx context.Context, id string) error {
	_, err := l.model.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("旅行攻略不存在")
	}
	return l.model.SoftDelete(ctx, id)
}

// GetPublicList 获取已发布旅行攻略列表（强制 status=2）。
func (l *TravelGuideLogic) GetPublicList(ctx context.Context, r *req.TravelGuideListReq) (*res.PageRes[res.TravelGuideRes], error) {
	published := int16(2)
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布旅行攻略详情，验证 status=2。
func (l *TravelGuideLogic) GetPublicDetail(ctx context.Context, id string) (*model.TravelGuide, error) {
	guide, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("旅行攻略不存在")
	}
	if guide.Status != 2 {
		return nil, fmt.Errorf("旅行攻略不存在")
	}
	return guide, nil
}

// GetHotList 获取热门旅行攻略列表。
func (l *TravelGuideLogic) GetHotList(ctx context.Context, count int) ([]model.TravelGuide, error) {
	return l.model.GetHotList(ctx, count)
}

// IncrementView 增加旅行攻略浏览量。
func (l *TravelGuideLogic) IncrementView(ctx context.Context, id string) error {
	return l.model.IncrementViewCount(ctx, id)
}

// IncrementLike 增加旅行攻略点赞数。
func (l *TravelGuideLogic) IncrementLike(ctx context.Context, id string) error {
	return l.model.IncrementLikeCount(ctx, id)
}

// toGuideRes 转换为列表响应。
func (l *TravelGuideLogic) toGuideRes(g *model.TravelGuide) res.TravelGuideRes {
	categoryID := ""
	if g.CategoryID != nil {
		categoryID = *g.CategoryID
	}
	categoryName := ""
	if g.Category != nil {
		categoryName = g.Category.Name
	}
	return res.TravelGuideRes{
		ID:           g.ID,
		Title:        g.Title,
		Summary:      g.Summary,
		CoverImage:   g.CoverImage,
		Status:       g.Status,
		Destination:  g.Destination,
		Region:       g.Region,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Days:         g.Days,
		BestMonth:    g.BestMonth,
		ViewCount:    g.ViewCount,
		LikeCount:    g.LikeCount,
		Rating:       g.Rating,
		ReviewCount:  g.ReviewCount,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

// toGuideDetailRes 转换为详情响应。
func (l *TravelGuideLogic) toGuideDetailRes(g *model.TravelGuide) *res.TravelGuideDetailRes {
	base := l.toGuideRes(g)
	return &res.TravelGuideDetailRes{
		TravelGuideRes: base,
		Attractions:    g.Attractions,
		Itinerary:      g.Itinerary,
		Reviews:        g.Reviews,
	}
}
