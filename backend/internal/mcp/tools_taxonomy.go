package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// resolveCategoryID 按名称解析分类 ID，不存在时自动创建（type=article）。名称为空返回 nil。
func (s *Service) resolveCategoryID(ctx context.Context, name string) (*string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	categories, err := s.categoryLogic.GetAll(ctx, "article")
	if err != nil {
		return nil, fmt.Errorf("查询分类列表失败: %w", err)
	}
	for _, c := range categories {
		if c.Name == name {
			id := c.ID
			return &id, nil
		}
	}
	created, err := s.categoryLogic.Create(ctx, &req.CreateCategoryReq{Name: name, Type: "article"})
	if err != nil {
		return nil, fmt.Errorf("创建分类 %q 失败: %w", name, err)
	}
	id := created.ID
	return &id, nil
}

// resolveTagIDs 按名称解析标签 ID 列表，不存在的自动创建；空入参返回空切片（清空标签）。
func (s *Service) resolveTagIDs(ctx context.Context, names []string) ([]string, error) {
	if len(names) == 0 {
		return []string{}, nil
	}
	tags, err := s.tagLogic.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询标签列表失败: %w", err)
	}
	byName := make(map[string]string, len(tags))
	for _, t := range tags {
		byName[t.Name] = t.ID
	}
	ids := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if id, ok := byName[name]; ok {
			ids = append(ids, id)
			continue
		}
		created, err := s.tagLogic.Create(ctx, &req.CreateTagReq{Name: name})
		if err != nil {
			return nil, fmt.Errorf("创建标签 %q 失败: %w", name, err)
		}
		ids = append(ids, created.ID)
	}
	return ids, nil
}

// CategoryItem 分类信息。
type CategoryItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ListCategoriesOut list_categories 出参。
type ListCategoriesOut struct {
	Items []CategoryItem `json:"items"`
}

// listCategories 查询文章分类列表。
func (s *Service) listCategories(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, ListCategoriesOut, error) {
	categories, err := s.categoryLogic.GetAll(ctx, "article")
	if err != nil {
		return nil, ListCategoriesOut{}, fmt.Errorf("查询分类列表失败: %w", err)
	}
	items := make([]CategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, CategoryItem{ID: c.ID, Name: c.Name, Slug: c.Slug})
	}
	return nil, ListCategoriesOut{Items: items}, nil
}

// TagItem 标签信息。
type TagItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListTagsOut list_tags 出参。
type ListTagsOut struct {
	Items []TagItem `json:"items"`
}

// listTags 查询标签列表。
func (s *Service) listTags(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, ListTagsOut, error) {
	tags, err := s.tagLogic.GetAll(ctx)
	if err != nil {
		return nil, ListTagsOut{}, fmt.Errorf("查询标签列表失败: %w", err)
	}
	items := make([]TagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, TagItem{ID: t.ID, Name: t.Name})
	}
	return nil, ListTagsOut{Items: items}, nil
}
