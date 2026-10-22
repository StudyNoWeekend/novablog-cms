package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/enum"
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
	Type string `json:"type"`
}

// ListCategoriesIn list_categories 入参。
type ListCategoriesIn struct {
	Type string `json:"type,omitempty" jsonschema:"按内容模块筛选分类，可选：article(默认)/travel/portfolio/music/video/book/recipe/game/project/open_source/tech_stack/fitness；不传返回文章分类"`
}

// ListCategoriesOut list_categories 出参。
type ListCategoriesOut struct {
	Items []CategoryItem `json:"items"`
}

// listCategories 查询分类列表（type 为空时默认 article）。
func (s *Service) listCategories(ctx context.Context, _ *sdk.CallToolRequest, in ListCategoriesIn) (*sdk.CallToolResult, ListCategoriesOut, error) {
	categoryType := strings.TrimSpace(in.Type)
	if categoryType == "" {
		categoryType = enum.CategoryTypeArticle
	}
	if !enum.IsValidCategoryType(categoryType) {
		return nil, ListCategoriesOut{}, fmt.Errorf("非法的分类类型 %q，可选：%s", categoryType, strings.Join(enum.CategoryTypes, "/"))
	}
	categories, err := s.categoryLogic.GetAll(ctx, categoryType)
	if err != nil {
		return nil, ListCategoriesOut{}, fmt.Errorf("查询分类列表失败: %w", err)
	}
	items := make([]CategoryItem, 0, len(categories))
	for _, c := range categories {
		items = append(items, CategoryItem{ID: c.ID, Name: c.Name, Slug: c.Slug, Type: c.Type})
	}
	return nil, ListCategoriesOut{Items: items}, nil
}

// CreateCategoryIn create_category 入参。
type CreateCategoryIn struct {
	Name        string `json:"name" jsonschema:"分类名称（必填，1-50 字）"`
	Slug        string `json:"slug,omitempty" jsonschema:"URL 别名，不传由后端生成"`
	Description string `json:"description,omitempty" jsonschema:"分类描述"`
	Type        string `json:"type,omitempty" jsonschema:"所属内容模块，可选：article(默认)/travel/portfolio/music/video/book/recipe/game/project/open_source/tech_stack/fitness"`
	SortOrder   int    `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// CreateCategoryOut create_category / update_category 出参。
type CreateCategoryOut struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// createCategory 创建分类。
func (s *Service) createCategory(ctx context.Context, _ *sdk.CallToolRequest, in CreateCategoryIn) (*sdk.CallToolResult, CreateCategoryOut, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, CreateCategoryOut{}, fmt.Errorf("name 不能为空")
	}
	categoryType := strings.TrimSpace(in.Type)
	if categoryType == "" {
		categoryType = enum.CategoryTypeArticle
	}
	if !enum.IsValidCategoryType(categoryType) {
		return nil, CreateCategoryOut{}, fmt.Errorf("非法的分类类型 %q，可选：%s", categoryType, strings.Join(enum.CategoryTypes, "/"))
	}
	created, err := s.categoryLogic.Create(ctx, &req.CreateCategoryReq{
		Name:        in.Name,
		Slug:        in.Slug,
		Description: in.Description,
		Type:        categoryType,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, CreateCategoryOut{}, fmt.Errorf("创建分类失败: %w", err)
	}
	return nil, CreateCategoryOut{ID: created.ID, Name: created.Name, Slug: created.Slug}, nil
}

// UpdateCategoryIn update_category 入参（指针字段=不传不修改）。
type UpdateCategoryIn struct {
	ID          string  `json:"id" jsonschema:"要更新的分类 ID"`
	Name        *string `json:"name,omitempty" jsonschema:"新名称"`
	Slug        *string `json:"slug,omitempty" jsonschema:"新 slug"`
	Description *string `json:"description,omitempty" jsonschema:"新描述"`
	SortOrder   *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateCategory 更新分类。
func (s *Service) updateCategory(ctx context.Context, _ *sdk.CallToolRequest, in UpdateCategoryIn) (*sdk.CallToolResult, CreateCategoryOut, error) {
	if in.ID == "" {
		return nil, CreateCategoryOut{}, fmt.Errorf("id 不能为空")
	}
	updated, err := s.categoryLogic.Update(ctx, in.ID, &req.UpdateCategoryReq{
		Name:        in.Name,
		Slug:        in.Slug,
		Description: in.Description,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, CreateCategoryOut{}, fmt.Errorf("更新分类失败: %w", err)
	}
	return nil, CreateCategoryOut{ID: updated.ID, Name: updated.Name, Slug: updated.Slug}, nil
}

// DeleteIn delete_category / delete_tag 等通用删除入参。
type DeleteIn struct {
	ID      string `json:"id" jsonschema:"要删除的记录 ID"`
	Confirm bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// DeleteOut 通用删除出参。
type DeleteOut struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// deleteCategory 删除分类（被内容占用时失败）。
func (s *Service) deleteCategory(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.categoryLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除分类失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
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

// CreateTagIn create_tag 入参。
type CreateTagIn struct {
	Name string `json:"name" jsonschema:"标签名称（必填，1-50 字）"`
}

// TagMutationOut create_tag / update_tag 出参。
type TagMutationOut struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// createTag 创建标签。
func (s *Service) createTag(ctx context.Context, _ *sdk.CallToolRequest, in CreateTagIn) (*sdk.CallToolResult, TagMutationOut, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, TagMutationOut{}, fmt.Errorf("name 不能为空")
	}
	created, err := s.tagLogic.Create(ctx, &req.CreateTagReq{Name: in.Name})
	if err != nil {
		return nil, TagMutationOut{}, fmt.Errorf("创建标签失败: %w", err)
	}
	return nil, TagMutationOut{ID: created.ID, Name: created.Name}, nil
}

// UpdateTagIn update_tag 入参。
type UpdateTagIn struct {
	ID   string `json:"id" jsonschema:"要重命名的标签 ID"`
	Name string `json:"name" jsonschema:"新标签名称（必填，1-50 字）"`
}

// updateTag 重命名标签。
func (s *Service) updateTag(ctx context.Context, _ *sdk.CallToolRequest, in UpdateTagIn) (*sdk.CallToolResult, TagMutationOut, error) {
	if in.ID == "" {
		return nil, TagMutationOut{}, fmt.Errorf("id 不能为空")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, TagMutationOut{}, fmt.Errorf("name 不能为空")
	}
	updated, err := s.tagLogic.Update(ctx, in.ID, &req.UpdateTagReq{Name: in.Name})
	if err != nil {
		return nil, TagMutationOut{}, fmt.Errorf("更新标签失败: %w", err)
	}
	return nil, TagMutationOut{ID: updated.ID, Name: updated.Name}, nil
}

// deleteTag 删除标签。
func (s *Service) deleteTag(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.tagLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除标签失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
