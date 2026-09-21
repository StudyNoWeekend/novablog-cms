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
	"gorm.io/gorm"
)

type ArticleLogic struct {
	model           *model.ArticleModel
	articleTagModel *model.ArticleTagModel
	categoryLogic   *CategoryLogic
	tagLogic        *TagLogic
}

func NewArticleLogic() *ArticleLogic {
	return &ArticleLogic{
		model:           model.NewArticle(),
		articleTagModel: model.NewArticleTag(),
		categoryLogic:   NewCategoryLogic(),
		tagLogic:        NewTagLogic(),
	}
}

// Create 创建文章（事务处理分类/标签关联）
func (l *ArticleLogic) Create(ctx context.Context, r *req.CreateArticleReq) (*res.ArticleDetailRes, error) {
	articleID := uuid.New().String()
	slug := generateSlug(r.Title)

	// 检查 slug 唯一性
	exists, err := l.model.ExistsBySlug(ctx, slug, "")
	if err != nil {
		return nil, fmt.Errorf("检查slug失败: %w", err)
	}
	if exists {
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	}

	// 默认值
	editorType := r.Type
	if editorType == 0 {
		editorType = 1
	}
	isTop := false
	if r.IsTop != nil {
		isTop = *r.IsTop
	}
	isComment := true
	if r.IsComment != nil {
		isComment = *r.IsComment
	}

	article := &model.Article{
		ID:         articleID,
		Title:      r.Title,
		Slug:       slug,
		Summary:    r.Summary,
		Content:    r.Content,
		CoverImage: r.CoverImage,
		CategoryID: r.CategoryID,
		Status:     r.Status,
		Type:       editorType,
		Extra:      r.Extra,
		IsTop:      isTop,
		IsComment:  isComment,
	}

	// 如果状态为已发布，设置发布时间
	if r.Status == 2 {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := l.model.Transaction(ctx, func(tx *gorm.DB) error {
		if err := l.model.CreateWithTx(ctx, tx, article); err != nil {
			return fmt.Errorf("创建文章失败: %w", err)
		}
		if err := l.model.ReplaceTagsWithTx(ctx, tx, articleID, r.TagIDs); err != nil {
			return fmt.Errorf("关联标签失败: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 重新查询完整信息
	return l.GetDetail(ctx, articleID)
}

// Update 更新文章
func (l *ArticleLogic) Update(ctx context.Context, id string, r *req.UpdateArticleReq) (*res.ArticleDetailRes, error) {
	article, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("文章不存在")
	}

	if r.Title != nil {
		article.Title = *r.Title
		slug := generateSlug(*r.Title)
		exists, _ := l.model.ExistsBySlug(ctx, slug, id)
		if !exists {
			article.Slug = slug
		}
	}
	if r.Content != nil {
		article.Content = *r.Content
	}
	if r.Summary != nil {
		article.Summary = *r.Summary
	}
	if r.CoverImage != nil {
		article.CoverImage = *r.CoverImage
	}
	if r.CategoryID != nil {
		article.CategoryID = r.CategoryID
	}
	if r.Type != nil {
		article.Type = *r.Type
	}
	if r.Extra != nil {
		article.Extra = r.Extra
	}
	if r.IsTop != nil {
		article.IsTop = *r.IsTop
	}
	if r.IsComment != nil {
		article.IsComment = *r.IsComment
	}

	// 处理状态变更
	if r.Status != nil {
		article.Status = *r.Status
		if *r.Status == 2 && article.PublishedAt == nil {
			now := time.Now()
			article.PublishedAt = &now
		}
	}

	if err := l.model.Transaction(ctx, func(tx *gorm.DB) error {
		if err := l.model.UpdateWithTx(ctx, tx, article); err != nil {
			return fmt.Errorf("更新文章失败: %w", err)
		}
		if r.TagIDs != nil {
			if err := l.model.ReplaceTagsWithTx(ctx, tx, id, r.TagIDs); err != nil {
				return fmt.Errorf("更新标签关联失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return l.GetDetail(ctx, id)
}

// GetList 获取文章列表（含分类/标签信息）
func (l *ArticleLogic) GetList(ctx context.Context, r *req.ArticleListReq) (*res.PageRes[res.ArticleRes], error) {
	articles, total, err := l.model.GetList(
		ctx, r.GetPage(), r.GetPageSize(),
		r.Status, r.CategoryID, r.Keyword,
	)
	if err != nil {
		return nil, err
	}

	var items []res.ArticleRes
	for _, a := range articles {
		item := l.toArticleRes(&a)
		items = append(items, item)
	}

	return res.NewPageRes(items, total, r.GetPage(), r.GetPageSize()), nil
}

// GetDetail 获取文章详情
func (l *ArticleLogic) GetDetail(ctx context.Context, id string) (*res.ArticleDetailRes, error) {
	article, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("文章不存在")
	}
	return l.toArticleDetail(article), nil
}

// GetBySlug 根据 slug 获取文章详情（公开）
func (l *ArticleLogic) GetBySlug(ctx context.Context, slug string) (*res.ArticleDetailRes, error) {
	article, err := l.model.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("文章不存在")
	}
	return l.toArticleDetail(article), nil
}

// UpdateStatus 更新文章状态
func (l *ArticleLogic) UpdateStatus(ctx context.Context, id string, r *req.UpdateStatusReq) error {
	_, err := l.model.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("文章不存在")
	}
	return l.model.UpdateStatus(ctx, id, r.Status)
}

// Delete 软删除文章
func (l *ArticleLogic) Delete(ctx context.Context, id string) error {
	_, err := l.model.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("文章不存在")
	}
	return l.model.SoftDelete(ctx, id)
}

// GetHotList 获取热门文章列表。
func (l *ArticleLogic) GetHotList(ctx context.Context, count int) ([]res.ArticleRes, error) {
	articles, err := l.model.GetHotList(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("查询热门文章失败: %w", err)
	}

	items := make([]res.ArticleRes, 0, len(articles))
	for i := range articles {
		items = append(items, l.toArticleRes(&articles[i]))
	}
	return items, nil
}

// GetRandomList 获取随机文章列表。
func (l *ArticleLogic) GetRandomList(ctx context.Context, count int) ([]res.ArticleRes, error) {
	articles, err := l.model.GetRandomList(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("查询随机文章失败: %w", err)
	}

	items := make([]res.ArticleRes, 0, len(articles))
	for i := range articles {
		items = append(items, l.toArticleRes(&articles[i]))
	}
	return items, nil
}

// IncrementView 增加文章浏览量。
func (l *ArticleLogic) IncrementView(ctx context.Context, slug string) error {
	return l.model.IncrementViewCount(ctx, slug)
}

// IncrementViewByID 根据 ID 增加文章浏览量（详情接口兜底计数用）。
func (l *ArticleLogic) IncrementViewByID(ctx context.Context, id string) error {
	return l.model.IncrementViewCountByID(ctx, id)
}

// toArticleRes 转换为列表响应
func (l *ArticleLogic) toArticleRes(a *model.Article) res.ArticleRes {
	var tagIDs []string
	var tagNames []string
	if a.Tags != nil {
		for _, t := range a.Tags {
			tagIDs = append(tagIDs, t.ID)
			tagNames = append(tagNames, t.Name)
		}
	}
	if tagIDs == nil {
		tagIDs = []string{}
	}
	if tagNames == nil {
		tagNames = []string{}
	}

	var categoryName string
	var categoryID string
	if a.Category != nil {
		categoryName = a.Category.Name
		categoryID = a.Category.ID
	} else if a.CategoryID != nil {
		categoryID = *a.CategoryID
	}

	return res.ArticleRes{
		ID:           a.ID,
		Title:        a.Title,
		Slug:         a.Slug,
		Summary:      a.Summary,
		CoverImage:   a.CoverImage,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		TagIDs:       tagIDs,
		TagNames:     tagNames,
		Status:       a.Status,
		Type:         a.Type,
		ViewCount:    a.ViewCount,
		CommentCount: a.CommentCount,
		IsTop:        a.IsTop,
		IsComment:    a.IsComment,
		PublishedAt:  a.PublishedAt,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// toArticleDetail 转换为详情响应
func (l *ArticleLogic) toArticleDetail(a *model.Article) *res.ArticleDetailRes {
	base := l.toArticleRes(a)
	return &res.ArticleDetailRes{
		ArticleRes: base,
		Content:    a.Content,
		Extra:      a.Extra,
	}
}

// generateSlug 从标题生成 slug（简单英文处理，中文转为拼音首字母或保留原样）
func generateSlug(title string) string {
	// 简单实现：小写+空格转横线+移除特殊字符
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除大部分特殊字符，保留中文
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r >= 0x4e00 {
			return r
		}
		return -1
	}, slug)
	if slug == "" {
		slug = uuid.New().String()[:8]
	}
	// 截断到 200
	runes := []rune(slug)
	if len(runes) > 200 {
		slug = string(runes[:200])
	}
	return strings.Trim(slug, "-")
}
