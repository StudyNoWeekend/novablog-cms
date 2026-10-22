package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTechStackTools 注册技术栈工具。
func (s *Service) registerTechStackTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_tech_stacks",
		Description: "分页查询技术栈条目列表（含完整字段），可按关键词、分类、掌握程度、状态过滤。",
	}, wrapTool(s, "list_tech_stacks", s.listTechStacks))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_tech_stack",
		Description: "按 ID 获取技术栈条目详情。",
	}, wrapTool(s, "get_tech_stack", s.getTechStack))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_tech_stack",
		Description: "新增技术栈条目。level: 1=了解 2=熟悉 3=熟练 4=精通；category_id 必须是已存在的技术栈分类。",
	}, wrapTool(s, "create_tech_stack", s.createTechStack))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_tech_stack",
		Description: "更新技术栈条目（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_tech_stack", s.updateTechStack))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_tech_stack",
		Description: "删除技术栈条目（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_tech_stack", s.deleteTechStack))
}

// ListTechStacksIn list_tech_stacks 入参。
type ListTechStacksIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按名称模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=tech_stack 的分类）"`
	Level      *int   `json:"level,omitempty" jsonschema:"按掌握程度筛选: 1=了解 2=熟悉 3=熟练 4=精通"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListTechStacksOut list_tech_stacks 出参。
type ListTechStacksOut struct {
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Items []res.TechStackItemRes `json:"items"`
}

// listTechStacks 分页查询技术栈列表。
func (s *Service) listTechStacks(ctx context.Context, _ *sdk.CallToolRequest, in ListTechStacksIn) (*sdk.CallToolResult, ListTechStacksOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.techStackLogic.GetList(ctx, &req.TechStackListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Level:      in.Level,
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListTechStacksOut{}, fmt.Errorf("查询技术栈列表失败: %w", err)
	}
	return nil, ListTechStacksOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetTechStackIn get_tech_stack 入参。
type GetTechStackIn struct {
	ID string `json:"id" jsonschema:"技术栈条目 ID"`
}

// getTechStack 获取技术栈详情。
func (s *Service) getTechStack(ctx context.Context, _ *sdk.CallToolRequest, in GetTechStackIn) (*sdk.CallToolResult, res.TechStackItemRes, error) {
	if in.ID == "" {
		return nil, res.TechStackItemRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.techStackLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.TechStackItemRes{}, fmt.Errorf("获取技术栈条目失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateTechStackIn create_tech_stack 入参。
type CreateTechStackIn struct {
	Name        string `json:"name" jsonschema:"技术名称（必填，1-100 字）"`
	CategoryID  string `json:"category_id,omitempty" jsonschema:"分类 ID（type=tech_stack 的分类，必须已存在）"`
	Icon        string `json:"icon,omitempty" jsonschema:"图标 URL，可先调用 upload_image 上传获取"`
	Level       *int   `json:"level,omitempty" jsonschema:"掌握程度: 1=了解 2=熟悉 3=熟练 4=精通"`
	Description string `json:"description,omitempty" jsonschema:"描述（≤500 字）"`
	Status      *int   `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int   `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createTechStack 新增技术栈条目。
func (s *Service) createTechStack(ctx context.Context, _ *sdk.CallToolRequest, in CreateTechStackIn) (*sdk.CallToolResult, res.TechStackItemRes, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, res.TechStackItemRes{}, fmt.Errorf("name 不能为空")
	}
	if in.Level != nil && (*in.Level < 1 || *in.Level > 4) {
		return nil, res.TechStackItemRes{}, fmt.Errorf("level 取值 1-4（1=了解 2=熟悉 3=熟练 4=精通）")
	}
	created, err := s.techStackLogic.Create(ctx, &req.CreateTechStackReq{
		Name:        in.Name,
		CategoryID:  stringPtr(in.CategoryID),
		Icon:        in.Icon,
		Level:       in.Level,
		Description: in.Description,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.TechStackItemRes{}, fmt.Errorf("创建技术栈条目失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateTechStackIn update_tech_stack 入参（指针字段=不传不修改）。
type UpdateTechStackIn struct {
	ID          string  `json:"id" jsonschema:"要更新的技术栈条目 ID"`
	Name        *string `json:"name,omitempty" jsonschema:"新名称"`
	CategoryID  *string `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Icon        *string `json:"icon,omitempty" jsonschema:"新图标 URL"`
	Level       *int    `json:"level,omitempty" jsonschema:"新掌握程度: 1=了解 2=熟悉 3=熟练 4=精通"`
	Description *string `json:"description,omitempty" jsonschema:"新描述"`
	Status      *int    `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder   *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateTechStack 更新技术栈条目。
func (s *Service) updateTechStack(ctx context.Context, _ *sdk.CallToolRequest, in UpdateTechStackIn) (*sdk.CallToolResult, res.TechStackItemRes, error) {
	if in.ID == "" {
		return nil, res.TechStackItemRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Level != nil && (*in.Level < 1 || *in.Level > 4) {
		return nil, res.TechStackItemRes{}, fmt.Errorf("level 取值 1-4（1=了解 2=熟悉 3=熟练 4=精通）")
	}
	updated, err := s.techStackLogic.Update(ctx, in.ID, &req.UpdateTechStackReq{
		Name:        in.Name,
		CategoryID:  in.CategoryID,
		Icon:        in.Icon,
		Level:       in.Level,
		Description: in.Description,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.TechStackItemRes{}, fmt.Errorf("更新技术栈条目失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteTechStack 删除技术栈条目。
func (s *Service) deleteTechStack(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.techStackLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除技术栈条目失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
