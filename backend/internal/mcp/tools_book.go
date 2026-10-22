package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerBookTools 注册读书书架工具。
func (s *Service) registerBookTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_books",
		Description: "分页查询读书书架列表（不含书评正文），可按关键词、阅读状态、分类、状态过滤。",
	}, wrapTool(s, "list_books", s.listBooks))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_book",
		Description: "按 ID 获取书籍详情（含完整书评）。",
	}, wrapTool(s, "get_book", s.getBook))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_book",
		Description: "新增书籍到书架。reading_status: want(想读)/reading(在读)/done(已读)；rating 0-5；started_at/finished_at 为 RFC3339 时间。",
	}, wrapTool(s, "create_book", s.createBook))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_book",
		Description: "更新书籍信息（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_book", s.updateBook))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_book",
		Description: "从书架删除书籍（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_book", s.deleteBook))
}

// ListBooksIn list_books 入参。
type ListBooksIn struct {
	Keyword       string `json:"keyword,omitempty" jsonschema:"按书名/作者模糊搜索"`
	ReadingStatus string `json:"reading_status,omitempty" jsonschema:"按阅读状态筛选: want/reading/done，不传返回全部"`
	CategoryID    string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=book 的分类）"`
	Status        *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page          int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize      int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListBooksOut list_books 出参。
type ListBooksOut struct {
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Items []res.BookCardRes `json:"items"`
}

// listBooks 分页查询书架列表。
func (s *Service) listBooks(ctx context.Context, _ *sdk.CallToolRequest, in ListBooksIn) (*sdk.CallToolResult, ListBooksOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.bookLogic.GetList(ctx, &req.BookListReq{
		PageReq:       req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:       stringPtr(in.Keyword),
		ReadingStatus: stringPtr(in.ReadingStatus),
		CategoryID:    stringPtr(in.CategoryID),
		Status:        in.Status,
	})
	if err != nil {
		return nil, ListBooksOut{}, fmt.Errorf("查询书架列表失败: %w", err)
	}
	return nil, ListBooksOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetBookIn get_book 入参。
type GetBookIn struct {
	ID string `json:"id" jsonschema:"书籍 ID"`
}

// getBook 获取书籍详情。
func (s *Service) getBook(ctx context.Context, _ *sdk.CallToolRequest, in GetBookIn) (*sdk.CallToolResult, res.BookRes, error) {
	if in.ID == "" {
		return nil, res.BookRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.bookLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.BookRes{}, fmt.Errorf("获取书籍失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateBookIn create_book 入参。
type CreateBookIn struct {
	Title         string     `json:"title" jsonschema:"书名（必填，1-255 字）"`
	Author        string     `json:"author,omitempty" jsonschema:"作者"`
	Cover         string     `json:"cover,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Rating        *int       `json:"rating,omitempty" jsonschema:"评分 0-5"`
	ReadingStatus string     `json:"reading_status,omitempty" jsonschema:"阅读状态: want(想读)/reading(在读)/done(已读)，默认 want"`
	Review        string     `json:"review,omitempty" jsonschema:"书评（Markdown，仅详情页展示）"`
	StartedAt     *time.Time `json:"started_at,omitempty" jsonschema:"开始阅读时间（RFC3339，如 2026-01-01T00:00:00Z）"`
	FinishedAt    *time.Time `json:"finished_at,omitempty" jsonschema:"读完时间（RFC3339）"`
	CategoryID    string     `json:"category_id,omitempty" jsonschema:"分类 ID（type=book 的分类）"`
	Status        *int       `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder     *int       `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createBook 新增书籍。
func (s *Service) createBook(ctx context.Context, _ *sdk.CallToolRequest, in CreateBookIn) (*sdk.CallToolResult, res.BookRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.BookRes{}, fmt.Errorf("title 不能为空")
	}
	if in.Rating != nil && (*in.Rating < 0 || *in.Rating > 5) {
		return nil, res.BookRes{}, fmt.Errorf("rating 取值 0-5")
	}
	switch in.ReadingStatus {
	case "", "want", "reading", "done":
	default:
		return nil, res.BookRes{}, fmt.Errorf("reading_status 取值非法：want / reading / done")
	}
	created, err := s.bookLogic.Create(ctx, &req.CreateBookReq{
		Title:         in.Title,
		Author:        in.Author,
		Cover:         in.Cover,
		Rating:        in.Rating,
		ReadingStatus: in.ReadingStatus,
		Review:        in.Review,
		StartedAt:     in.StartedAt,
		FinishedAt:    in.FinishedAt,
		CategoryID:    stringPtr(in.CategoryID),
		Status:        in.Status,
		SortOrder:     in.SortOrder,
	})
	if err != nil {
		return nil, res.BookRes{}, fmt.Errorf("创建书籍失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateBookIn update_book 入参（指针字段=不传不修改）。
type UpdateBookIn struct {
	ID            string     `json:"id" jsonschema:"要更新的书籍 ID"`
	Title         *string    `json:"title,omitempty" jsonschema:"新书名"`
	Author        *string    `json:"author,omitempty" jsonschema:"新作者"`
	Cover         *string    `json:"cover,omitempty" jsonschema:"新封面图 URL"`
	Rating        *int       `json:"rating,omitempty" jsonschema:"新评分 0-5"`
	ReadingStatus *string    `json:"reading_status,omitempty" jsonschema:"新阅读状态: want/reading/done"`
	Review        *string    `json:"review,omitempty" jsonschema:"新书评（Markdown）"`
	StartedAt     *time.Time `json:"started_at,omitempty" jsonschema:"新开始阅读时间（RFC3339）"`
	FinishedAt    *time.Time `json:"finished_at,omitempty" jsonschema:"新读完时间（RFC3339）"`
	CategoryID    *string    `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Status        *int       `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder     *int       `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateBook 更新书籍。
func (s *Service) updateBook(ctx context.Context, _ *sdk.CallToolRequest, in UpdateBookIn) (*sdk.CallToolResult, res.BookRes, error) {
	if in.ID == "" {
		return nil, res.BookRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Rating != nil && (*in.Rating < 0 || *in.Rating > 5) {
		return nil, res.BookRes{}, fmt.Errorf("rating 取值 0-5")
	}
	if in.ReadingStatus != nil {
		switch *in.ReadingStatus {
		case "want", "reading", "done":
		default:
			return nil, res.BookRes{}, fmt.Errorf("reading_status 取值非法：want / reading / done")
		}
	}
	updated, err := s.bookLogic.Update(ctx, in.ID, &req.UpdateBookReq{
		Title:         in.Title,
		Author:        in.Author,
		Cover:         in.Cover,
		Rating:        in.Rating,
		ReadingStatus: in.ReadingStatus,
		Review:        in.Review,
		StartedAt:     in.StartedAt,
		FinishedAt:    in.FinishedAt,
		CategoryID:    in.CategoryID,
		Status:        in.Status,
		SortOrder:     in.SortOrder,
	})
	if err != nil {
		return nil, res.BookRes{}, fmt.Errorf("更新书籍失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteBook 删除书籍。
func (s *Service) deleteBook(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.bookLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除书籍失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
