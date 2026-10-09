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

// GameLogic 游戏库业务逻辑结构体。
type GameLogic struct {
	gameModel *model.GameModel
}

// NewGameLogic 创建 GameLogic 实例。
func NewGameLogic() *GameLogic {
	return &GameLogic{gameModel: model.NewGame()}
}

// Create 创建游戏。
func (l *GameLogic) Create(ctx context.Context, r *req.CreateGameReq) (*res.GameRes, error) {
	m := &model.Game{
		ID:          uuid.New().String(),
		Title:       strings.TrimSpace(r.Title),
		Cover:       r.Cover,
		Platform:    r.Platform,
		Genre:       r.Genre,
		PlayStatus:  "want",
		ShortReview: r.ShortReview,
	}
	if r.PlayStatus != "" {
		m.PlayStatus = r.PlayStatus
	}
	if r.PlayHours != nil {
		m.PlayHours = *r.PlayHours
	}
	if r.Rating != nil {
		m.Rating = *r.Rating
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.gameModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建游戏失败: %w", err)
	}
	return l.toRes(m), nil
}

// GetList 分页查询游戏列表。
func (l *GameLogic) GetList(ctx context.Context, r *req.GameListReq) (*res.GameListRes, error) {
	list, total, err := l.gameModel.GetList(ctx, r.Keyword, r.Platform, r.Genre, r.PlayStatus, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询游戏列表失败: %w", err)
	}

	items := make([]res.GameCardRes, 0, len(list))
	for i := range list {
		items = append(items, l.toCard(&list[i]))
	}
	return &res.GameListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询游戏详情。
func (l *GameLogic) GetByID(ctx context.Context, id string) (*res.GameRes, error) {
	m, err := l.gameModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("游戏不存在")
	}
	return l.toRes(m), nil
}

// Update 更新游戏（部分更新语义）。
func (l *GameLogic) Update(ctx context.Context, id string, r *req.UpdateGameReq) (*res.GameRes, error) {
	m, err := l.gameModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("游戏不存在")
	}

	if r.Title != nil {
		m.Title = strings.TrimSpace(*r.Title)
	}
	if r.Cover != nil {
		m.Cover = *r.Cover
	}
	if r.Platform != nil {
		m.Platform = *r.Platform
	}
	if r.Genre != nil {
		m.Genre = *r.Genre
	}
	if r.PlayStatus != nil {
		m.PlayStatus = *r.PlayStatus
	}
	if r.PlayHours != nil {
		m.PlayHours = *r.PlayHours
	}
	if r.Rating != nil {
		m.Rating = *r.Rating
	}
	if r.ShortReview != nil {
		m.ShortReview = *r.ShortReview
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.gameModel.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新游戏失败: %w", err)
	}
	return l.toRes(m), nil
}

// Delete 删除游戏。
func (l *GameLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.gameModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("游戏不存在")
	}
	if err := l.gameModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除游戏失败: %w", err)
	}
	return nil
}

// GetPublicList 获取已发布游戏列表（强制 status=1）。
func (l *GameLogic) GetPublicList(ctx context.Context, r *req.GameListReq) (*res.GameListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布游戏详情（未发布视为不存在）。
func (l *GameLogic) GetPublicDetail(ctx context.Context, id string) (*res.GameRes, error) {
	m, err := l.gameModel.GetByID(ctx, id)
	if err != nil || m.Status != 1 {
		return nil, fmt.Errorf("游戏不存在")
	}
	return l.toRes(m), nil
}

// toRes 转换为游戏详情响应。
func (l *GameLogic) toRes(m *model.Game) *res.GameRes {
	return &res.GameRes{
		ID:          m.ID,
		Title:       m.Title,
		Cover:       m.Cover,
		Platform:    m.Platform,
		Genre:       m.Genre,
		PlayStatus:  m.PlayStatus,
		PlayHours:   m.PlayHours,
		Rating:      m.Rating,
		ShortReview: m.ShortReview,
		Status:      m.Status,
		SortOrder:   m.SortOrder,
		ViewCount:   int64(m.ViewCount),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toCard 转换为游戏列表卡片响应。
func (l *GameLogic) toCard(m *model.Game) res.GameCardRes {
	return res.GameCardRes{
		ID:         m.ID,
		Title:      m.Title,
		Cover:      m.Cover,
		Platform:   m.Platform,
		Genre:      m.Genre,
		PlayStatus: m.PlayStatus,
		PlayHours:  m.PlayHours,
		Rating:     m.Rating,
		Status:     m.Status,
		SortOrder:  m.SortOrder,
		ViewCount:  int64(m.ViewCount),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
