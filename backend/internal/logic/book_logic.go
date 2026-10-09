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

// BookLogic 读书书架业务逻辑结构体。
type BookLogic struct {
	bookModel *model.BookModel
}

// NewBookLogic 创建 BookLogic 实例。
func NewBookLogic() *BookLogic {
	return &BookLogic{bookModel: model.NewBook()}
}

// Create 创建书籍。
func (l *BookLogic) Create(ctx context.Context, r *req.CreateBookReq) (*res.BookRes, error) {
	m := &model.Book{
		ID:            uuid.New().String(),
		Title:         strings.TrimSpace(r.Title),
		Author:        r.Author,
		Cover:         r.Cover,
		ReadingStatus: "want",
		Review:        r.Review,
		StartedAt:     r.StartedAt,
		FinishedAt:    r.FinishedAt,
	}
	if r.Rating != nil {
		m.Rating = *r.Rating
	}
	if r.ReadingStatus != "" {
		m.ReadingStatus = r.ReadingStatus
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.bookModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建书籍失败: %w", err)
	}
	return l.toRes(m), nil
}

// GetList 分页查询书籍列表。
func (l *BookLogic) GetList(ctx context.Context, r *req.BookListReq) (*res.BookListRes, error) {
	list, total, err := l.bookModel.GetList(ctx, r.Keyword, r.ReadingStatus, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询书籍列表失败: %w", err)
	}

	items := make([]res.BookCardRes, 0, len(list))
	for i := range list {
		items = append(items, l.toCard(&list[i]))
	}
	return &res.BookListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询书籍详情。
func (l *BookLogic) GetByID(ctx context.Context, id string) (*res.BookRes, error) {
	m, err := l.bookModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("书籍不存在")
	}
	return l.toRes(m), nil
}

// Update 更新书籍（部分更新语义）。
func (l *BookLogic) Update(ctx context.Context, id string, r *req.UpdateBookReq) (*res.BookRes, error) {
	m, err := l.bookModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("书籍不存在")
	}

	if r.Title != nil {
		m.Title = strings.TrimSpace(*r.Title)
	}
	if r.Author != nil {
		m.Author = *r.Author
	}
	if r.Cover != nil {
		m.Cover = *r.Cover
	}
	if r.Rating != nil {
		m.Rating = *r.Rating
	}
	if r.ReadingStatus != nil {
		m.ReadingStatus = *r.ReadingStatus
	}
	if r.Review != nil {
		m.Review = *r.Review
	}
	if r.StartedAt != nil {
		m.StartedAt = r.StartedAt
	}
	if r.FinishedAt != nil {
		m.FinishedAt = r.FinishedAt
	}
	if r.Status != nil {
		m.Status = *r.Status
	}
	if r.SortOrder != nil {
		m.SortOrder = *r.SortOrder
	}

	if err := l.bookModel.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("更新书籍失败: %w", err)
	}
	return l.toRes(m), nil
}

// Delete 删除书籍。
func (l *BookLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.bookModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("书籍不存在")
	}
	if err := l.bookModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除书籍失败: %w", err)
	}
	return nil
}

// GetPublicList 获取已发布书籍列表（强制 status=1）。
func (l *BookLogic) GetPublicList(ctx context.Context, r *req.BookListReq) (*res.BookListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布书籍详情（未发布视为不存在）。
func (l *BookLogic) GetPublicDetail(ctx context.Context, id string) (*res.BookRes, error) {
	m, err := l.bookModel.GetByID(ctx, id)
	if err != nil || m.Status != 1 {
		return nil, fmt.Errorf("书籍不存在")
	}
	return l.toRes(m), nil
}

// toRes 转换为书籍详情响应。
func (l *BookLogic) toRes(m *model.Book) *res.BookRes {
	return &res.BookRes{
		ID:            m.ID,
		Title:         m.Title,
		Author:        m.Author,
		Cover:         m.Cover,
		Rating:        m.Rating,
		ReadingStatus: m.ReadingStatus,
		Review:        m.Review,
		StartedAt:     m.StartedAt,
		FinishedAt:    m.FinishedAt,
		Status:        m.Status,
		SortOrder:     m.SortOrder,
		ViewCount:     int64(m.ViewCount),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// toCard 转换为书籍列表卡片响应。
func (l *BookLogic) toCard(m *model.Book) res.BookCardRes {
	return res.BookCardRes{
		ID:            m.ID,
		Title:         m.Title,
		Author:        m.Author,
		Cover:         m.Cover,
		Rating:        m.Rating,
		ReadingStatus: m.ReadingStatus,
		Status:        m.Status,
		SortOrder:     m.SortOrder,
		ViewCount:     int64(m.ViewCount),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
