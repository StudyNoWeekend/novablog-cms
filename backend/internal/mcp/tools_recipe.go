package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerRecipeTools 注册美食菜谱工具。
func (s *Service) registerRecipeTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_recipes",
		Description: "分页查询美食菜谱列表（不含食材/步骤明细），可按关键词、难度、分类、状态过滤。",
	}, wrapTool(s, "list_recipes", s.listRecipes))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_recipe",
		Description: "按 ID 获取菜谱详情（含食材清单与烹饪步骤）。",
	}, wrapTool(s, "get_recipe", s.getRecipe))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_recipe",
		Description: "新增美食菜谱。title 必填；ingredients/steps 为对象数组（如 [{\"name\":\"鸡蛋\",\"amount\":\"2个\"}] / [{\"text\":\"打散\",\"image\":\"\"}]）；difficulty: 1=简单 2=中等 3=困难。",
	}, wrapTool(s, "create_recipe", s.createRecipe))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_recipe",
		Description: "更新菜谱（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_recipe", s.updateRecipe))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_recipe",
		Description: "删除菜谱（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_recipe", s.deleteRecipe))
}

// ListRecipesIn list_recipes 入参。
type ListRecipesIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题/简介/标签模糊搜索"`
	Difficulty *int   `json:"difficulty,omitempty" jsonschema:"按难度筛选: 1=简单 2=中等 3=困难"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=recipe 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListRecipesOut list_recipes 出参。
type ListRecipesOut struct {
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Items []res.RecipeCardRes `json:"items"`
}

// listRecipes 分页查询菜谱列表。
func (s *Service) listRecipes(ctx context.Context, _ *sdk.CallToolRequest, in ListRecipesIn) (*sdk.CallToolResult, ListRecipesOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.recipeLogic.GetList(ctx, &req.RecipeListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		Difficulty: in.Difficulty,
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListRecipesOut{}, fmt.Errorf("查询菜谱列表失败: %w", err)
	}
	return nil, ListRecipesOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetRecipeIn get_recipe 入参。
type GetRecipeIn struct {
	ID string `json:"id" jsonschema:"菜谱 ID"`
}

// getRecipe 获取菜谱详情。
func (s *Service) getRecipe(ctx context.Context, _ *sdk.CallToolRequest, in GetRecipeIn) (*sdk.CallToolResult, res.RecipeRes, error) {
	if in.ID == "" {
		return nil, res.RecipeRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.recipeLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.RecipeRes{}, fmt.Errorf("获取菜谱失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateRecipeIn create_recipe 入参。
type CreateRecipeIn struct {
	Title       string           `json:"title" jsonschema:"菜谱标题（必填，1-255 字）"`
	Cover       string           `json:"cover,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Summary     string           `json:"summary,omitempty" jsonschema:"菜谱简介（≤500 字）"`
	Ingredients []map[string]any `json:"ingredients,omitempty" jsonschema:"食材清单，对象数组整体写入，如 [{\"name\":\"鸡蛋\",\"amount\":\"2个\"}]"`
	Steps       []map[string]any `json:"steps,omitempty" jsonschema:"烹饪步骤，对象数组整体写入，如 [{\"text\":\"打散鸡蛋\",\"image\":\"\"}]"`
	Difficulty  *int             `json:"difficulty,omitempty" jsonschema:"难度: 1=简单 2=中等 3=困难"`
	Minutes     *int             `json:"minutes,omitempty" jsonschema:"耗时（分钟，≥0）"`
	Servings    *int             `json:"servings,omitempty" jsonschema:"份量（人份，≥1）"`
	Tags        string           `json:"tags,omitempty" jsonschema:"标签，逗号分隔（如 家常,快手）"`
	CategoryID  string           `json:"category_id,omitempty" jsonschema:"分类 ID（type=recipe 的分类）"`
	Status      *int             `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int             `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createRecipe 新增菜谱。
func (s *Service) createRecipe(ctx context.Context, _ *sdk.CallToolRequest, in CreateRecipeIn) (*sdk.CallToolResult, res.RecipeRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.RecipeRes{}, fmt.Errorf("title 不能为空")
	}
	if in.Difficulty != nil && (*in.Difficulty < 1 || *in.Difficulty > 3) {
		return nil, res.RecipeRes{}, fmt.Errorf("difficulty 取值 1-3（1=简单 2=中等 3=困难）")
	}
	created, err := s.recipeLogic.Create(ctx, &req.CreateRecipeReq{
		Title:       in.Title,
		Cover:       in.Cover,
		Summary:     in.Summary,
		Ingredients: in.Ingredients,
		Steps:       in.Steps,
		Difficulty:  in.Difficulty,
		Minutes:     in.Minutes,
		Servings:    in.Servings,
		Tags:        in.Tags,
		CategoryID:  stringPtr(in.CategoryID),
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.RecipeRes{}, fmt.Errorf("创建菜谱失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateRecipeIn update_recipe 入参（指针字段=不传不修改）。
type UpdateRecipeIn struct {
	ID          string           `json:"id" jsonschema:"要更新的菜谱 ID"`
	Title       *string          `json:"title,omitempty" jsonschema:"新标题"`
	Cover       *string          `json:"cover,omitempty" jsonschema:"新封面图 URL"`
	Summary     *string          `json:"summary,omitempty" jsonschema:"新简介"`
	Ingredients []map[string]any `json:"ingredients,omitempty" jsonschema:"替换后的完整食材清单；不传该字段则不修改"`
	Steps       []map[string]any `json:"steps,omitempty" jsonschema:"替换后的完整步骤列表；不传该字段则不修改"`
	Difficulty  *int             `json:"difficulty,omitempty" jsonschema:"新难度: 1=简单 2=中等 3=困难"`
	Minutes     *int             `json:"minutes,omitempty" jsonschema:"新耗时（分钟）"`
	Servings    *int             `json:"servings,omitempty" jsonschema:"新份量（人份）"`
	Tags        *string          `json:"tags,omitempty" jsonschema:"新标签，逗号分隔"`
	CategoryID  *string          `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Status      *int             `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder   *int             `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateRecipe 更新菜谱。
func (s *Service) updateRecipe(ctx context.Context, _ *sdk.CallToolRequest, in UpdateRecipeIn) (*sdk.CallToolResult, res.RecipeRes, error) {
	if in.ID == "" {
		return nil, res.RecipeRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Difficulty != nil && (*in.Difficulty < 1 || *in.Difficulty > 3) {
		return nil, res.RecipeRes{}, fmt.Errorf("difficulty 取值 1-3（1=简单 2=中等 3=困难）")
	}
	updated, err := s.recipeLogic.Update(ctx, in.ID, &req.UpdateRecipeReq{
		Title:       in.Title,
		Cover:       in.Cover,
		Summary:     in.Summary,
		Ingredients: in.Ingredients,
		Steps:       in.Steps,
		Difficulty:  in.Difficulty,
		Minutes:     in.Minutes,
		Servings:    in.Servings,
		Tags:        in.Tags,
		CategoryID:  in.CategoryID,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.RecipeRes{}, fmt.Errorf("更新菜谱失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteRecipe 删除菜谱。
func (s *Service) deleteRecipe(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.recipeLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除菜谱失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
