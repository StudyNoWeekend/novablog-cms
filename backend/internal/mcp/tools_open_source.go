package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerOpenSourceTools 注册开源作品工具。
func (s *Service) registerOpenSourceTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_open_sources",
		Description: "分页查询开源作品列表（不含 README 正文），可按关键词、分类、状态过滤。",
	}, wrapTool(s, "list_open_sources", s.listOpenSources))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_open_source",
		Description: "按 ID 获取开源作品详情（含 README 原文与 Star 数）。",
	}, wrapTool(s, "get_open_source", s.getOpenSource))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_open_source",
		Description: "新增开源作品。repo_url 必须是合法的 GitHub 仓库地址；默认自动拉取 README 与 Stars（fetch_remote=false 可跳过）。",
	}, wrapTool(s, "create_open_source", s.createOpenSource))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_open_source",
		Description: "更新开源作品（整体替换传入了的字段，未传入的保持不变；repo_url 变更会自动重新拉取远端数据）。",
	}, wrapTool(s, "update_open_source", s.updateOpenSource))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_open_source",
		Description: "删除开源作品（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_open_source", s.deleteOpenSource))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "refresh_open_source",
		Description: "手动刷新开源作品：重新拉取 GitHub README 与仓库元数据（Stars 等）。",
	}, wrapTool(s, "refresh_open_source", s.refreshOpenSource))
}

// ListOpenSourcesIn list_open_sources 入参。
type ListOpenSourcesIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按名称/简介模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=open_source 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListOpenSourcesOut list_open_sources 出参。
type ListOpenSourcesOut struct {
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Items []res.OpenSourceCardRes `json:"items"`
}

// listOpenSources 分页查询开源作品列表。
func (s *Service) listOpenSources(ctx context.Context, _ *sdk.CallToolRequest, in ListOpenSourcesIn) (*sdk.CallToolResult, ListOpenSourcesOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.openSourceLogic.GetList(ctx, &req.OpenSourceListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListOpenSourcesOut{}, fmt.Errorf("查询开源作品列表失败: %w", err)
	}
	return nil, ListOpenSourcesOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetOpenSourceIn get_open_source 入参。
type GetOpenSourceIn struct {
	ID string `json:"id" jsonschema:"开源作品 ID"`
}

// getOpenSource 获取开源作品详情。
func (s *Service) getOpenSource(ctx context.Context, _ *sdk.CallToolRequest, in GetOpenSourceIn) (*sdk.CallToolResult, res.OpenSourceRes, error) {
	if in.ID == "" {
		return nil, res.OpenSourceRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.openSourceLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.OpenSourceRes{}, fmt.Errorf("获取开源作品失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateOpenSourceIn create_open_source 入参。
type CreateOpenSourceIn struct {
	Name        string `json:"name" jsonschema:"作品名称（必填，1-255 字）"`
	RepoURL     string `json:"repo_url" jsonschema:"GitHub 仓库地址（必填，如 https://github.com/owner/repo）"`
	Summary     string `json:"summary,omitempty" jsonschema:"一句话简介（≤500 字）"`
	Language    string `json:"language,omitempty" jsonschema:"主要编程语言"`
	Topics      string `json:"topics,omitempty" jsonschema:"主题标签，逗号分隔"`
	Homepage    string `json:"homepage,omitempty" jsonschema:"项目主页地址"`
	CategoryID  string `json:"category_id,omitempty" jsonschema:"分类 ID（type=open_source 的分类）"`
	Status      *int   `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int   `json:"sort_order,omitempty" jsonschema:"排序权重"`
	FetchRemote *bool  `json:"fetch_remote,omitempty" jsonschema:"是否自动拉取 GitHub 元数据与 README，默认 true"`
}

// createOpenSource 新增开源作品。
func (s *Service) createOpenSource(ctx context.Context, _ *sdk.CallToolRequest, in CreateOpenSourceIn) (*sdk.CallToolResult, res.OpenSourceRes, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.RepoURL) == "" {
		return nil, res.OpenSourceRes{}, fmt.Errorf("name 与 repo_url 为必填项")
	}
	created, err := s.openSourceLogic.Create(ctx, &req.CreateOpenSourceReq{
		Name:        in.Name,
		RepoURL:     in.RepoURL,
		Summary:     in.Summary,
		Language:    in.Language,
		Topics:      in.Topics,
		Homepage:    in.Homepage,
		CategoryID:  stringPtr(in.CategoryID),
		Status:      in.Status,
		SortOrder:   in.SortOrder,
		FetchRemote: in.FetchRemote,
	})
	if err != nil {
		return nil, res.OpenSourceRes{}, fmt.Errorf("创建开源作品失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateOpenSourceIn update_open_source 入参（指针字段=不传不修改）。
type UpdateOpenSourceIn struct {
	ID         string  `json:"id" jsonschema:"要更新的开源作品 ID"`
	Name       *string `json:"name,omitempty" jsonschema:"新名称"`
	RepoURL    *string `json:"repo_url,omitempty" jsonschema:"新 GitHub 仓库地址（变更后自动重新拉取）"`
	Summary    *string `json:"summary,omitempty" jsonschema:"新简介"`
	Language   *string `json:"language,omitempty" jsonschema:"新主要语言"`
	Topics     *string `json:"topics,omitempty" jsonschema:"新主题标签，逗号分隔"`
	Homepage   *string `json:"homepage,omitempty" jsonschema:"新主页地址"`
	CategoryID *string `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Status     *int    `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder  *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateOpenSource 更新开源作品。
func (s *Service) updateOpenSource(ctx context.Context, _ *sdk.CallToolRequest, in UpdateOpenSourceIn) (*sdk.CallToolResult, res.OpenSourceRes, error) {
	if in.ID == "" {
		return nil, res.OpenSourceRes{}, fmt.Errorf("id 不能为空")
	}
	updated, err := s.openSourceLogic.Update(ctx, in.ID, &req.UpdateOpenSourceReq{
		Name:       in.Name,
		RepoURL:    in.RepoURL,
		Summary:    in.Summary,
		Language:   in.Language,
		Topics:     in.Topics,
		Homepage:   in.Homepage,
		CategoryID: in.CategoryID,
		Status:     in.Status,
		SortOrder:  in.SortOrder,
	})
	if err != nil {
		return nil, res.OpenSourceRes{}, fmt.Errorf("更新开源作品失败: %w", err)
	}
	return nil, *updated, nil
}

// refreshOpenSource 刷新开源作品远端数据。
func (s *Service) refreshOpenSource(ctx context.Context, _ *sdk.CallToolRequest, in GetOpenSourceIn) (*sdk.CallToolResult, res.OpenSourceRes, error) {
	if in.ID == "" {
		return nil, res.OpenSourceRes{}, fmt.Errorf("id 不能为空")
	}
	refreshed, err := s.openSourceLogic.Refresh(ctx, in.ID)
	if err != nil {
		return nil, res.OpenSourceRes{}, fmt.Errorf("刷新开源作品失败: %w", err)
	}
	return nil, *refreshed, nil
}

// deleteOpenSource 删除开源作品。
func (s *Service) deleteOpenSource(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.openSourceLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除开源作品失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
