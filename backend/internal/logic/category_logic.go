package logic

import (
	"context"
	"fmt"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

type CategoryLogic struct {
	model *model.CategoryModel
}

func NewCategoryLogic() *CategoryLogic {
	return &CategoryLogic{model: model.NewCategory()}
}

func (l *CategoryLogic) Create(ctx context.Context, r *req.CreateCategoryReq) (*res.CategoryRes, error) {
	// 默认类型为 article；type 必须是登记过的合法值
	categoryType := r.Type
	if categoryType == "" {
		categoryType = enum.CategoryTypeArticle
	}
	if !enum.IsValidCategoryType(categoryType) {
		return nil, fmt.Errorf("分类类型不合法")
	}

	// 检查名称是否重复
	existing, _ := l.model.GetByName(ctx, r.Name, categoryType)
	if existing != nil {
		return nil, fmt.Errorf("分类名称已存在")
	}

	slug := r.Slug
	if slug == "" {
		slug = r.Name // 简单处理，实际可用拼音库
	}

	cat := &model.Category{
		ID:          uuid.New().String(),
		Name:        r.Name,
		Slug:        slug,
		Description: r.Description,
		Type:        categoryType,
		SortOrder:   r.SortOrder,
	}

	if err := l.model.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	return &res.CategoryRes{
		ID:          cat.ID,
		Name:        cat.Name,
		Slug:        cat.Slug,
		Description: cat.Description,
		Type:        cat.Type,
		SortOrder:   cat.SortOrder,
		CreatedAt:   cat.CreatedAt,
	}, nil
}

func (l *CategoryLogic) GetAll(ctx context.Context, categoryType string) ([]res.CategoryRes, error) {
	categories, err := l.model.GetAll(ctx, categoryType)
	if err != nil {
		return nil, err
	}
	var result []res.CategoryRes
	for _, c := range categories {
		result = append(result, res.CategoryRes{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			Type:        c.Type,
			SortOrder:   c.SortOrder,
			CreatedAt:   c.CreatedAt,
		})
	}
	if result == nil {
		result = []res.CategoryRes{}
	}
	return result, nil
}

func (l *CategoryLogic) GetByID(ctx context.Context, id string) (*res.CategoryRes, error) {
	c, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("分类不存在")
	}
	return &res.CategoryRes{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Type:        c.Type,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
	}, nil
}

func (l *CategoryLogic) Update(ctx context.Context, id string, r *req.UpdateCategoryReq) (*res.CategoryRes, error) {
	c, err := l.model.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	if r.Name != nil {
		existing, _ := l.model.GetByName(ctx, *r.Name, c.Type)
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("分类名称已存在")
		}
		c.Name = *r.Name
	}
	if r.Slug != nil {
		c.Slug = *r.Slug
	}
	if r.Description != nil {
		c.Description = *r.Description
	}
	if r.Description != nil {
		c.Description = *r.Description
	}
	if r.Type != nil {
		if !enum.IsValidCategoryType(*r.Type) {
			return nil, fmt.Errorf("分类类型不合法")
		}
		c.Type = *r.Type
	}
	if r.SortOrder != nil {
		c.SortOrder = *r.SortOrder
	}

	if err := l.model.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("更新分类失败: %w", err)
	}

	return &res.CategoryRes{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Type:        c.Type,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
	}, nil
}

// categoryIDOf 将可空分类 ID 转为响应字符串，空返回空串。
func categoryIDOf(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

// categoryNameOf 从预加载的关联分类取名称，未加载返回空串。
func categoryNameOf(c *model.Category) string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (l *CategoryLogic) Delete(ctx context.Context, id string) error {
	hasArticles, err := l.model.HasArticles(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasArticles {
		return fmt.Errorf("该分类下存在文章，无法删除")
	}
	hasSongs, err := l.model.HasSongs(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasSongs {
		return fmt.Errorf("该分类下存在歌曲，无法删除")
	}
	hasPortfolios, err := l.model.HasPortfolios(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasPortfolios {
		return fmt.Errorf("该分类下存在作品集，无法删除")
	}
	hasTravelGuides, err := l.model.HasTravelGuides(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasTravelGuides {
		return fmt.Errorf("该分类下存在旅行攻略，无法删除")
	}
	hasVideos, err := l.model.HasVideos(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasVideos {
		return fmt.Errorf("该分类下存在视频作品，无法删除")
	}
	hasBooks, err := l.model.HasBooks(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasBooks {
		return fmt.Errorf("该分类下存在书籍，无法删除")
	}
	hasRecipes, err := l.model.HasRecipes(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasRecipes {
		return fmt.Errorf("该分类下存在菜谱，无法删除")
	}
	hasGames, err := l.model.HasGames(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasGames {
		return fmt.Errorf("该分类下存在游戏，无法删除")
	}
	hasProjects, err := l.model.HasProjects(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasProjects {
		return fmt.Errorf("该分类下存在项目经历，无法删除")
	}
	hasOpenSourceWorks, err := l.model.HasOpenSourceWorks(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasOpenSourceWorks {
		return fmt.Errorf("该分类下存在开源作品，无法删除")
	}
	hasTechStackItems, err := l.model.HasTechStackItems(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasTechStackItems {
		return fmt.Errorf("该分类下存在技术栈条目，无法删除")
	}
	hasFitnessRecords, err := l.model.HasFitnessRecords(ctx, id)
	if err != nil {
		return fmt.Errorf("检查分类关联失败: %w", err)
	}
	if hasFitnessRecords {
		return fmt.Errorf("该分类下存在健身训练记录，无法删除")
	}
	return l.model.Delete(ctx, id)
}
