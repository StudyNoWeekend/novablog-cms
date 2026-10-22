package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPortfolioTools 注册摄影作品集工具。
func (s *Service) registerPortfolioTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_portfolios",
		Description: "分页查询摄影作品集列表（含封面与作品数统计，不含作品明细），可按关键词、分类、状态过滤。",
	}, wrapTool(s, "list_portfolios", s.listPortfolios))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_portfolio",
		Description: "按 ID 获取作品集详情（含全部作品项）。",
	}, wrapTool(s, "get_portfolio", s.getPortfolio))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_portfolio",
		Description: "新增摄影作品集。cover_mode=1（独立封面）时 cover_preset_id 必填且需为已存在的媒体预设。",
	}, wrapTool(s, "create_portfolio", s.createPortfolio))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_portfolio",
		Description: "更新作品集（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_portfolio", s.updatePortfolio))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_portfolio",
		Description: "删除作品集（级联软删其下全部作品项）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_portfolio", s.deletePortfolio))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "add_portfolio_item",
		Description: "向作品集添加一个作品项（preset_id 为媒体库预设 ID，可先 list_media 查询）。",
	}, wrapTool(s, "add_portfolio_item", s.addPortfolioItem))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_portfolio_item",
		Description: "更新作品集内的单个作品项（整体替换传入了的字段）。",
	}, wrapTool(s, "update_portfolio_item", s.updatePortfolioItem))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_portfolio_item",
		Description: "删除作品集内的单个作品项。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_portfolio_item", s.deletePortfolioItem))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "sort_portfolio_items",
		Description: "批量调整作品集内作品的展示顺序。",
	}, wrapTool(s, "sort_portfolio_items", s.sortPortfolioItems))
}

// ListPortfoliosIn list_portfolios 入参。
type ListPortfoliosIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按名称模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=portfolio 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListPortfoliosOut list_portfolios 出参。
type ListPortfoliosOut struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Items []res.PortfolioRes `json:"items"`
}

// listPortfolios 分页查询作品集列表。
func (s *Service) listPortfolios(ctx context.Context, _ *sdk.CallToolRequest, in ListPortfoliosIn) (*sdk.CallToolResult, ListPortfoliosOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.portfolioLogic.GetList(ctx, &req.PortfolioListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListPortfoliosOut{}, fmt.Errorf("查询作品集列表失败: %w", err)
	}
	return nil, ListPortfoliosOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetPortfolioIn get_portfolio 入参。
type GetPortfolioIn struct {
	ID string `json:"id" jsonschema:"作品集 ID"`
}

// getPortfolio 获取作品集详情（含作品项列表）。
func (s *Service) getPortfolio(ctx context.Context, _ *sdk.CallToolRequest, in GetPortfolioIn) (*sdk.CallToolResult, res.PortfolioDetailRes, error) {
	if in.ID == "" {
		return nil, res.PortfolioDetailRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.portfolioLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.PortfolioDetailRes{}, fmt.Errorf("获取作品集失败: %w", err)
	}
	return nil, *detail, nil
}

// CreatePortfolioIn create_portfolio 入参。
type CreatePortfolioIn struct {
	Name          string `json:"name" jsonschema:"作品集名称（必填，1-255 字）"`
	Description   string `json:"description,omitempty" jsonschema:"作品集描述"`
	CoverMode     *int   `json:"cover_mode,omitempty" jsonschema:"封面模式: 0=使用排序第一的作品(默认) 1=独立设置封面"`
	CoverPresetID string `json:"cover_preset_id,omitempty" jsonschema:"独立封面预设 ID（cover_mode=1 时必填）"`
	Status        *int   `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder     *int   `json:"sort_order,omitempty" jsonschema:"排序权重"`
	CategoryID    string `json:"category_id,omitempty" jsonschema:"分类 ID（type=portfolio 的分类）"`
}

// createPortfolio 新增作品集。
func (s *Service) createPortfolio(ctx context.Context, _ *sdk.CallToolRequest, in CreatePortfolioIn) (*sdk.CallToolResult, res.PortfolioRes, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, res.PortfolioRes{}, fmt.Errorf("name 不能为空")
	}
	if in.CoverMode != nil && *in.CoverMode == 1 && strings.TrimSpace(in.CoverPresetID) == "" {
		return nil, res.PortfolioRes{}, fmt.Errorf("cover_mode=1 时必须提供 cover_preset_id")
	}
	var coverMode *int
	if in.CoverMode != nil && (*in.CoverMode == 0 || *in.CoverMode == 1) {
		coverMode = in.CoverMode
	}
	created, err := s.portfolioLogic.Create(ctx, &req.CreatePortfolioReq{
		Name:          in.Name,
		Description:   in.Description,
		CoverMode:     coverMode,
		CoverPresetID: stringPtr(in.CoverPresetID),
		Status:        in.Status,
		SortOrder:     in.SortOrder,
		CategoryID:    stringPtr(in.CategoryID),
	})
	if err != nil {
		return nil, res.PortfolioRes{}, fmt.Errorf("创建作品集失败: %w", err)
	}
	return nil, *created, nil
}

// UpdatePortfolioIn update_portfolio 入参（指针字段=不传不修改）。
type UpdatePortfolioIn struct {
	ID            string  `json:"id" jsonschema:"要更新的作品集 ID"`
	Name          *string `json:"name,omitempty" jsonschema:"新名称"`
	Description   *string `json:"description,omitempty" jsonschema:"新描述"`
	CoverMode     *int    `json:"cover_mode,omitempty" jsonschema:"新封面模式: 0=使用排序第一的作品 1=独立设置封面"`
	CoverPresetID *string `json:"cover_preset_id,omitempty" jsonschema:"新独立封面预设 ID（cover_mode=1 时必填）"`
	Status        *int    `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder     *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
	CategoryID    *string `json:"category_id,omitempty" jsonschema:"新分类 ID"`
}

// updatePortfolio 更新作品集。
func (s *Service) updatePortfolio(ctx context.Context, _ *sdk.CallToolRequest, in UpdatePortfolioIn) (*sdk.CallToolResult, res.PortfolioRes, error) {
	if in.ID == "" {
		return nil, res.PortfolioRes{}, fmt.Errorf("id 不能为空")
	}
	updated, err := s.portfolioLogic.Update(ctx, in.ID, &req.UpdatePortfolioReq{
		Name:          in.Name,
		Description:   in.Description,
		CoverMode:     in.CoverMode,
		CoverPresetID: in.CoverPresetID,
		Status:        in.Status,
		SortOrder:     in.SortOrder,
		CategoryID:    in.CategoryID,
	})
	if err != nil {
		return nil, res.PortfolioRes{}, fmt.Errorf("更新作品集失败: %w", err)
	}
	return nil, *updated, nil
}

// PortfolioItemIn 作品项入参。
type PortfolioItemIn struct {
	PresetID    string `json:"preset_id" jsonschema:"媒体预设 ID（必填，来自媒体库预设）"`
	Title       string `json:"title" jsonschema:"作品标题（必填，1-255 字）"`
	Description string `json:"description,omitempty" jsonschema:"作品描述"`
}

// AddPortfolioItemIn add_portfolio_item 入参。
type AddPortfolioItemIn struct {
	PortfolioID string `json:"portfolio_id" jsonschema:"目标作品集 ID"`
	PresetID    string `json:"preset_id" jsonschema:"媒体预设 ID（必填）"`
	Title       string `json:"title" jsonschema:"作品标题（必填）"`
	Description string `json:"description,omitempty" jsonschema:"作品描述"`
}

// PortfolioItemOut 作品项出参。
type PortfolioItemOut struct {
	ID          string `json:"id"`
	PortfolioID string `json:"portfolio_id"`
	PresetID    string `json:"preset_id"`
	Title       string `json:"title"`
	SortOrder   int    `json:"sort_order"`
	OutputURL   string `json:"output_url"`
}

// addPortfolioItem 添加作品项。
func (s *Service) addPortfolioItem(ctx context.Context, _ *sdk.CallToolRequest, in AddPortfolioItemIn) (*sdk.CallToolResult, res.PortfolioItemRes, error) {
	if in.PortfolioID == "" {
		return nil, res.PortfolioItemRes{}, fmt.Errorf("portfolio_id 不能为空")
	}
	if strings.TrimSpace(in.PresetID) == "" || strings.TrimSpace(in.Title) == "" {
		return nil, res.PortfolioItemRes{}, fmt.Errorf("preset_id 与 title 为必填项")
	}
	item, err := s.portfolioLogic.AddItem(ctx, in.PortfolioID, &req.CreatePortfolioItemReq{
		PresetID:    in.PresetID,
		Title:       in.Title,
		Description: in.Description,
	})
	if err != nil {
		return nil, res.PortfolioItemRes{}, fmt.Errorf("添加作品失败: %w", err)
	}
	return nil, *item, nil
}

// UpdatePortfolioItemIn update_portfolio_item 入参（指针字段=不传不修改）。
type UpdatePortfolioItemIn struct {
	PortfolioID string  `json:"portfolio_id" jsonschema:"所属作品集 ID"`
	ItemID      string  `json:"item_id" jsonschema:"要更新的作品项 ID"`
	PresetID    *string `json:"preset_id,omitempty" jsonschema:"新预设 ID"`
	Title       *string `json:"title,omitempty" jsonschema:"新标题"`
	Description *string `json:"description,omitempty" jsonschema:"新描述"`
}

// updatePortfolioItem 更新作品项。
func (s *Service) updatePortfolioItem(ctx context.Context, _ *sdk.CallToolRequest, in UpdatePortfolioItemIn) (*sdk.CallToolResult, res.PortfolioItemRes, error) {
	if in.PortfolioID == "" || in.ItemID == "" {
		return nil, res.PortfolioItemRes{}, fmt.Errorf("portfolio_id 与 item_id 不能为空")
	}
	updated, err := s.portfolioLogic.UpdateItem(ctx, in.PortfolioID, in.ItemID, &req.UpdatePortfolioItemReq{
		PresetID:    in.PresetID,
		Title:       in.Title,
		Description: in.Description,
	})
	if err != nil {
		return nil, res.PortfolioItemRes{}, fmt.Errorf("更新作品项失败: %w", err)
	}
	return nil, *updated, nil
}

// DeletePortfolioItemIn delete_portfolio_item 入参。
type DeletePortfolioItemIn struct {
	PortfolioID string `json:"portfolio_id" jsonschema:"所属作品集 ID"`
	ItemID      string `json:"item_id" jsonschema:"要删除的作品项 ID"`
	Confirm     bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// deletePortfolioItem 删除作品项。
func (s *Service) deletePortfolioItem(ctx context.Context, _ *sdk.CallToolRequest, in DeletePortfolioItemIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.PortfolioID == "" || in.ItemID == "" {
		return nil, DeleteOut{}, fmt.Errorf("portfolio_id 与 item_id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.portfolioLogic.DeleteItem(ctx, in.PortfolioID, in.ItemID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除作品项失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ItemID}, nil
}

// SortPortfolioItemsIn sort_portfolio_items 入参。
type SortPortfolioItemsIn struct {
	PortfolioID string       `json:"portfolio_id" jsonschema:"作品集 ID"`
	Items       []SortItemIn `json:"items" jsonschema:"排序项数组，如 [{\"id\":\"...\",\"sort_order\":1}]，数字越小越靠前"`
}

// SortItemIn 单个排序项。
type SortItemIn struct {
	ID        string `json:"id" jsonschema:"作品项 ID"`
	SortOrder int    `json:"sort_order" jsonschema:"排序权重，数字越小越靠前"`
}

// SortPortfoliosOut sort_portfolio_items 出参。
type SortPortfoliosOut struct {
	Sorted bool   `json:"sorted"`
	Count  int    `json:"count"`
	ID     string `json:"id"`
}

// sortPortfolioItems 批量排序作品项。
func (s *Service) sortPortfolioItems(ctx context.Context, _ *sdk.CallToolRequest, in SortPortfolioItemsIn) (*sdk.CallToolResult, SortPortfoliosOut, error) {
	if in.PortfolioID == "" {
		return nil, SortPortfoliosOut{}, fmt.Errorf("portfolio_id 不能为空")
	}
	if len(in.Items) == 0 {
		return nil, SortPortfoliosOut{}, fmt.Errorf("items 不能为空")
	}
	items := make([]req.SortPortfolioItemReq, 0, len(in.Items))
	for _, it := range in.Items {
		items = append(items, req.SortPortfolioItemReq{ID: it.ID, SortOrder: it.SortOrder})
	}
	if err := s.portfolioLogic.SortItems(ctx, in.PortfolioID, &req.SortPortfolioItemsReq{Items: items}); err != nil {
		return nil, SortPortfoliosOut{}, fmt.Errorf("排序作品项失败: %w", err)
	}
	return nil, SortPortfoliosOut{Sorted: true, Count: len(in.Items), ID: in.PortfolioID}, nil
}

// DeletePortfolioIn delete_portfolio 入参。
type DeletePortfolioIn struct {
	ID      string `json:"id" jsonschema:"要删除的作品集 ID"`
	Confirm bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// deletePortfolio 删除作品集（级联软删其下全部作品项）。
func (s *Service) deletePortfolio(ctx context.Context, _ *sdk.CallToolRequest, in DeletePortfolioIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作（会同时删除该作品集下的全部作品项），请确认后显式传 confirm=true 再执行")
	}
	if err := s.portfolioLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除作品集失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
