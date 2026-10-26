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
	coverMediaID, coverPresetID, cover, err := normalizeMediaRef(ctx, r.CoverMediaID, r.CoverPresetID, r.Cover)
	if err != nil {
		return nil, err
	}
	m := &model.Book{
		ID:            uuid.New().String(),
		Title:         strings.TrimSpace(r.Title),
		Author:        r.Author,
		Cover:         cover,
		CoverMediaID:  mediaRefPtr(coverMediaID),
		CoverPresetID: mediaRefPtr(coverPresetID),
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
	if r.CategoryID != nil {
		m.CategoryID = r.CategoryID
	}

	if err := l.bookModel.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("创建书籍失败: %w", err)
	}
	return l.toRes(m), nil
}

// GetList 分页查询书籍列表。
func (l *BookLogic) GetList(ctx context.Context, r *req.BookListReq) (*res.BookListRes, error) {
	list, total, err := l.bookModel.GetList(ctx, r.Keyword, r.ReadingStatus, r.CategoryID, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询书籍列表失败: %w", err)
	}

	items := make([]res.BookCardRes, 0, len(list))
	l.resolveCoverURLs(ctx, list)
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
	l.resolveCoverURL(ctx, m)
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
	if r.Cover != nil || r.CoverMediaID != nil {
		mediaID, presetID := "", ""
		if r.CoverMediaID != nil {
			mediaID = *r.CoverMediaID
		}
		if r.CoverPresetID != nil {
			presetID = *r.CoverPresetID
		}
		coverURL := ""
		if r.Cover != nil {
			coverURL = *r.Cover
		}
		var coverMediaID, coverPresetID, cover string
		coverMediaID, coverPresetID, cover, err = normalizeMediaRef(ctx, mediaID, presetID, coverURL)
		if err != nil {
			return nil, err
		}
		m.Cover = cover
		m.CoverMediaID = mediaRefPtr(coverMediaID)
		m.CoverPresetID = mediaRefPtr(coverPresetID)
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
	if r.CategoryID != nil {
		m.CategoryID = r.CategoryID
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
	l.resolveCoverURL(ctx, m)
	return l.toRes(m), nil
}

// resolveCoverURLs 批量解析书籍封面媒体引用为当前完整 URL（外链行保持原值）。
func (l *BookLogic) resolveCoverURLs(ctx context.Context, books []model.Book) {
	ptrs := make([]*model.Book, len(books))
	for i := range books {
		ptrs[i] = &books[i]
	}
	model.ResolveCoverURLs(ctx, ptrs,
		func(b *model.Book) (string, string) { return refStr(b.CoverMediaID), refStr(b.CoverPresetID) },
		func(b *model.Book, u string) { b.Cover = u },
	)
}

// resolveCoverURL 解析单本书籍封面媒体引用。
func (l *BookLogic) resolveCoverURL(ctx context.Context, b *model.Book) {
	model.ResolveCoverURLs(ctx, []*model.Book{b},
		func(x *model.Book) (string, string) { return refStr(x.CoverMediaID), refStr(x.CoverPresetID) },
		func(x *model.Book, u string) { x.Cover = u },
	)
}

// toRes 转换为书籍详情响应。
func (l *BookLogic) toRes(m *model.Book) *res.BookRes {
	return &res.BookRes{
		ID:            m.ID,
		Title:         m.Title,
		Author:        m.Author,
		Cover:         m.Cover,
		CoverMediaID:  refStr(m.CoverMediaID),
		CoverPresetID: refStr(m.CoverPresetID),
		Rating:        m.Rating,
		ReadingStatus: m.ReadingStatus,
		Review:        m.Review,
		StartedAt:     m.StartedAt,
		FinishedAt:    m.FinishedAt,
		Status:        m.Status,
		SortOrder:     m.SortOrder,
		CategoryID:    categoryIDOf(m.CategoryID),
		CategoryName:  categoryNameOf(m.Category),
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
		CategoryID:    categoryIDOf(m.CategoryID),
		CategoryName:  categoryNameOf(m.Category),
		ViewCount:     int64(m.ViewCount),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
