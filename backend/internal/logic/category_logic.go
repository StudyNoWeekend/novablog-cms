package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"

	"github.com/google/uuid"
)

type CategoryLogic struct {
	model *model.CategoryModel
}

func NewCategoryLogic() *CategoryLogic {
	return &CategoryLogic{model: model.NewCategory()}
}

func (l *CategoryLogic) Create(ctx context.Context, r *req.CreateCategoryReq) (*res.CategoryRes, error) {
	// 默认类型为 article
	categoryType := r.Type
	if categoryType == "" {
		categoryType = "article"
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
	if r.Type != nil {
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
	return l.model.Delete(ctx, id)
}
