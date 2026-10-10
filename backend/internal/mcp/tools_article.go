package mcp

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ArticleSummary 文章摘要信息（列表用，不含正文）。
type ArticleSummary struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Status      string     `json:"status" jsonschema:"文章状态: draft/published/offline"`
	Category    string     `json:"category"`
	Tags        []string   `json:"tags"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListArticlesIn list_articles 入参。
type ListArticlesIn struct {
	Keyword  string `json:"keyword,omitempty" jsonschema:"按标题模糊搜索的关键词"`
	Status   string `json:"status,omitempty" jsonschema:"按状态筛选: draft(草稿)/published(已发布)/offline(已下架)，不传返回全部"`
	Page     int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListArticlesOut list_articles 出参。
type ListArticlesOut struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Items []ArticleSummary `json:"items"`
}

// listArticles 分页查询文章列表。
func (s *Service) listArticles(ctx context.Context, _ *sdk.CallToolRequest, in ListArticlesIn) (*sdk.CallToolResult, ListArticlesOut, error) {
	status, err := statusFromString(in.Status)
	if err != nil {
		return nil, ListArticlesOut{}, err
	}
	var statusPtr *int16
	if status != 0 {
		statusPtr = &status
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}

	page, err := s.articleLogic.GetList(ctx, &req.ArticleListReq{
		PageReq: req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Status:  statusPtr,
		Keyword: &in.Keyword,
	})
	if err != nil {
		return nil, ListArticlesOut{}, fmt.Errorf("查询文章列表失败: %w", err)
	}

	items := make([]ArticleSummary, 0, len(page.List))
	for _, a := range page.List {
		items = append(items, ArticleSummary{
			ID:          a.ID,
			Slug:        a.Slug,
			Title:       a.Title,
			Summary:     a.Summary,
			Status:      statusToString(a.Status),
			Category:    a.CategoryName,
			Tags:        a.TagNames,
			PublishedAt: a.PublishedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}
	return nil, ListArticlesOut{Total: page.Total, Page: page.Page, Items: items}, nil
}

// GetArticleIn get_article 入参。
type GetArticleIn struct {
	IDOrSlug string `json:"id_or_slug" jsonschema:"文章 ID 或 slug"`
}

// GetArticleOut get_article 出参。
type GetArticleOut struct {
	ArticleSummary
	Content string `json:"content" jsonschema:"Markdown 正文"`
}

// getArticle 获取文章完整内容。
func (s *Service) getArticle(ctx context.Context, _ *sdk.CallToolRequest, in GetArticleIn) (*sdk.CallToolResult, GetArticleOut, error) {
	if in.IDOrSlug == "" {
		return nil, GetArticleOut{}, fmt.Errorf("id_or_slug 不能为空")
	}
	detail, err := s.resolveArticleDetail(ctx, in.IDOrSlug)
	if err != nil {
		return nil, GetArticleOut{}, err
	}
	return nil, GetArticleOut{
		ArticleSummary: ArticleSummary{
			ID:          detail.ID,
			Slug:        detail.Slug,
			Title:       detail.Title,
			Summary:     detail.Summary,
			Status:      statusToString(detail.Status),
			Category:    detail.CategoryName,
			Tags:        detail.TagNames,
			PublishedAt: detail.PublishedAt,
			UpdatedAt:   detail.UpdatedAt,
		},
		Content: detail.Content,
	}, nil
}

// CreateArticleIn create_article 入参。
type CreateArticleIn struct {
	Title         string   `json:"title" jsonschema:"文章标题（必填）"`
	Content       string   `json:"content" jsonschema:"文章正文，Markdown 格式（必填）"`
	Summary       string   `json:"summary,omitempty" jsonschema:"文章摘要，不传则自动从正文提取"`
	TagNames      []string `json:"tag_names,omitempty" jsonschema:"标签名列表，不存在的标签会自动创建"`
	CategoryName  string   `json:"category_name,omitempty" jsonschema:"分类名，不存在的分类会自动创建"`
	CoverImageURL string   `json:"cover_image_url,omitempty" jsonschema:"封面图片 URL，可先调用 upload_image 上传获取"`
	Status        string   `json:"status,omitempty" jsonschema:"draft(草稿，默认) / published(立即发布)"`
}

// ArticleMutationOut 创建/更新/状态变更的统一出参。
type ArticleMutationOut struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// createArticle 创建文章。
func (s *Service) createArticle(ctx context.Context, _ *sdk.CallToolRequest, in CreateArticleIn) (*sdk.CallToolResult, ArticleMutationOut, error) {
	if in.Title == "" || in.Content == "" {
		return nil, ArticleMutationOut{}, fmt.Errorf("title 和 content 为必填项")
	}
	status, err := statusFromString(in.Status)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}
	if status == 0 {
		status = 1
	}

	categoryID, err := s.resolveCategoryID(ctx, in.CategoryName)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}
	tagIDs, err := s.resolveTagIDs(ctx, in.TagNames)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}

	summary := in.Summary
	if summary == "" {
		summary = truncateSummary(in.Content)
	}

	detail, err := s.articleLogic.Create(ctx, &req.CreateArticleReq{
		Title:      in.Title,
		Content:    in.Content,
		Summary:    summary,
		CoverImage: in.CoverImageURL,
		CategoryID: categoryID,
		TagIDs:     tagIDs,
		Status:     status,
		Type:       1, // Markdown
	})
	if err != nil {
		return nil, ArticleMutationOut{}, fmt.Errorf("创建文章失败: %w", err)
	}
	return nil, mutationOut(detail), nil
}

// UpdateArticleIn update_article 入参。
type UpdateArticleIn struct {
	IDOrSlug      string   `json:"id_or_slug" jsonschema:"要更新的文章 ID 或 slug"`
	Title         *string  `json:"title,omitempty" jsonschema:"新标题"`
	Content       *string  `json:"content,omitempty" jsonschema:"新正文（Markdown），整体替换"`
	Summary       *string  `json:"summary,omitempty" jsonschema:"新摘要"`
	TagNames      []string `json:"tag_names,omitempty" jsonschema:"替换后的完整标签名列表；不传该字段则不修改标签，传空数组则清空标签"`
	CategoryName  *string  `json:"category_name,omitempty" jsonschema:"新分类名"`
	CoverImageURL *string  `json:"cover_image_url,omitempty" jsonschema:"新封面图 URL"`
	Status        *string  `json:"status,omitempty" jsonschema:"新状态: draft/published/offline"`
}

// updateArticle 更新文章。
func (s *Service) updateArticle(ctx context.Context, _ *sdk.CallToolRequest, in UpdateArticleIn) (*sdk.CallToolResult, ArticleMutationOut, error) {
	if in.IDOrSlug == "" {
		return nil, ArticleMutationOut{}, fmt.Errorf("id_or_slug 不能为空")
	}
	current, err := s.resolveArticleDetail(ctx, in.IDOrSlug)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}

	patch := req.UpdateArticleReq{
		Title:      in.Title,
		Content:    in.Content,
		Summary:    in.Summary,
		CoverImage: in.CoverImageURL,
	}
	if in.Status != nil {
		status, err := statusFromString(*in.Status)
		if err != nil {
			return nil, ArticleMutationOut{}, err
		}
		if status != 0 {
			patch.Status = &status
		}
	}
	if in.CategoryName != nil {
		categoryID, err := s.resolveCategoryID(ctx, *in.CategoryName)
		if err != nil {
			return nil, ArticleMutationOut{}, err
		}
		patch.CategoryID = categoryID
	}
	if in.TagNames != nil {
		tagIDs, err := s.resolveTagIDs(ctx, in.TagNames)
		if err != nil {
			return nil, ArticleMutationOut{}, err
		}
		patch.TagIDs = tagIDs
	}

	detail, err := s.articleLogic.Update(ctx, current.ID, &patch)
	if err != nil {
		return nil, ArticleMutationOut{}, fmt.Errorf("更新文章失败: %w", err)
	}
	return nil, mutationOut(detail), nil
}

// SetArticleStatusIn set_article_status 入参。
type SetArticleStatusIn struct {
	IDOrSlug string `json:"id_or_slug" jsonschema:"文章 ID 或 slug"`
	Action   string `json:"action" jsonschema:"目标状态: draft(转草稿) / publish(发布) / offline(下架)"`
}

// setArticleStatus 修改文章状态。
func (s *Service) setArticleStatus(ctx context.Context, _ *sdk.CallToolRequest, in SetArticleStatusIn) (*sdk.CallToolResult, ArticleMutationOut, error) {
	if in.IDOrSlug == "" {
		return nil, ArticleMutationOut{}, fmt.Errorf("id_or_slug 不能为空")
	}
	status, err := statusFromString(in.Action)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}
	if status == 0 {
		return nil, ArticleMutationOut{}, fmt.Errorf("action 不能为空，可选：draft / publish / offline")
	}
	current, err := s.resolveArticleDetail(ctx, in.IDOrSlug)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}
	if err := s.articleLogic.UpdateStatus(ctx, current.ID, &req.UpdateStatusReq{Status: status}); err != nil {
		return nil, ArticleMutationOut{}, fmt.Errorf("更新文章状态失败: %w", err)
	}
	detail, err := s.articleLogic.GetDetail(ctx, current.ID)
	if err != nil {
		return nil, ArticleMutationOut{}, err
	}
	return nil, mutationOut(detail), nil
}

// DeleteArticleIn delete_article 入参。
type DeleteArticleIn struct {
	IDOrSlug string `json:"id_or_slug" jsonschema:"要删除的文章 ID 或 slug"`
	Confirm  bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// DeleteArticleOut delete_article 出参。
type DeleteArticleOut struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// deleteArticle 删除文章（软删除）。
func (s *Service) deleteArticle(ctx context.Context, _ *sdk.CallToolRequest, in DeleteArticleIn) (*sdk.CallToolResult, DeleteArticleOut, error) {
	if in.IDOrSlug == "" {
		return nil, DeleteArticleOut{}, fmt.Errorf("id_or_slug 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteArticleOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	current, err := s.resolveArticleDetail(ctx, in.IDOrSlug)
	if err != nil {
		return nil, DeleteArticleOut{}, err
	}
	if err := s.articleLogic.Delete(ctx, current.ID); err != nil {
		return nil, DeleteArticleOut{}, fmt.Errorf("删除文章失败: %w", err)
	}
	return nil, DeleteArticleOut{Deleted: true, ID: current.ID}, nil
}

// resolveArticleDetail 按 ID 或 slug 解析文章详情。
func (s *Service) resolveArticleDetail(ctx context.Context, idOrSlug string) (*res.ArticleDetailRes, error) {
	detail, err := s.articleLogic.GetDetail(ctx, idOrSlug)
	if err == nil {
		return detail, nil
	}
	detail, err = s.articleLogic.GetBySlug(ctx, idOrSlug)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %s", idOrSlug)
	}
	return detail, nil
}

// mutationOut 转换文章详情为变更结果。
func mutationOut(detail *res.ArticleDetailRes) ArticleMutationOut {
	return ArticleMutationOut{
		ID:          detail.ID,
		Slug:        detail.Slug,
		Title:       detail.Title,
		Status:      statusToString(detail.Status),
		PublishedAt: detail.PublishedAt,
	}
}
