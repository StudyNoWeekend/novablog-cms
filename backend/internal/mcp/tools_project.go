package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerProjectTools 注册项目经历工具。
func (s *Service) registerProjectTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_projects",
		Description: "分页查询项目经历列表（含完整字段），可按关键词、分类、状态过滤。",
	}, wrapTool(s, "list_projects", s.listProjects))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_project",
		Description: "按 ID 获取项目经历详情。",
	}, wrapTool(s, "get_project", s.getProject))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_project",
		Description: "新增项目经历。start_date/end_date 为 YYYY-MM 格式（end_date 空表示至今）；status: 0=草稿(默认) 1=已发布。",
	}, wrapTool(s, "create_project", s.createProject))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_project",
		Description: "更新项目经历（整体替换传入了的字段，未传入的保持不变；日期字段传空串表示清空）。",
	}, wrapTool(s, "update_project", s.updateProject))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_project",
		Description: "删除项目经历（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_project", s.deleteProject))
}

// ListProjectsIn list_projects 入参。
type ListProjectsIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题/简介模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=project 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListProjectsOut list_projects 出参。
type ListProjectsOut struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Items []res.ProjectRes `json:"items"`
}

// listProjects 分页查询项目列表。
func (s *Service) listProjects(ctx context.Context, _ *sdk.CallToolRequest, in ListProjectsIn) (*sdk.CallToolResult, ListProjectsOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.projectLogic.GetList(ctx, &req.ProjectListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListProjectsOut{}, fmt.Errorf("查询项目列表失败: %w", err)
	}
	return nil, ListProjectsOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetProjectIn get_project 入参。
type GetProjectIn struct {
	ID string `json:"id" jsonschema:"项目 ID"`
}

// getProject 获取项目详情。
func (s *Service) getProject(ctx context.Context, _ *sdk.CallToolRequest, in GetProjectIn) (*sdk.CallToolResult, res.ProjectRes, error) {
	if in.ID == "" {
		return nil, res.ProjectRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.projectLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.ProjectRes{}, fmt.Errorf("获取项目失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateProjectIn create_project 入参。
type CreateProjectIn struct {
	Title       string `json:"title" jsonschema:"项目名称（必填，1-255 字）"`
	CategoryID  string `json:"category_id,omitempty" jsonschema:"分类 ID（type=project 的分类，可先 list_categories 查询）"`
	Role        string `json:"role,omitempty" jsonschema:"担任角色"`
	Client      string `json:"client,omitempty" jsonschema:"客户/甲方"`
	CoverURL    string `json:"cover_url,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Summary     string `json:"summary,omitempty" jsonschema:"项目简介（≤500 字）"`
	Description string `json:"description,omitempty" jsonschema:"项目详述（Markdown）"`
	TechStack   string `json:"tech_stack,omitempty" jsonschema:"技能/工具标签，逗号分隔"`
	StartDate   string `json:"start_date,omitempty" jsonschema:"开始月份，格式 YYYY-MM"`
	EndDate     string `json:"end_date,omitempty" jsonschema:"结束月份，格式 YYYY-MM；不传表示至今"`
	ProjectURL  string `json:"project_url,omitempty" jsonschema:"项目线上地址"`
	RepoURL     string `json:"repo_url,omitempty" jsonschema:"代码仓库地址"`
	Status      *int   `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int   `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createProject 新增项目经历。
func (s *Service) createProject(ctx context.Context, _ *sdk.CallToolRequest, in CreateProjectIn) (*sdk.CallToolResult, res.ProjectRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.ProjectRes{}, fmt.Errorf("title 不能为空")
	}
	if err := validateMonth(in.StartDate, "start_date"); err != nil {
		return nil, res.ProjectRes{}, err
	}
	if err := validateMonth(in.EndDate, "end_date"); err != nil {
		return nil, res.ProjectRes{}, err
	}
	created, err := s.projectLogic.Create(ctx, &req.CreateProjectReq{
		Title:       in.Title,
		CategoryID:  stringPtr(in.CategoryID),
		Role:        in.Role,
		Client:      in.Client,
		CoverURL:    in.CoverURL,
		Summary:     in.Summary,
		Description: in.Description,
		TechStack:   in.TechStack,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		ProjectURL:  in.ProjectURL,
		RepoURL:     in.RepoURL,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.ProjectRes{}, fmt.Errorf("创建项目失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateProjectIn update_project 入参（指针字段=不传不修改；日期传空串表示清空）。
type UpdateProjectIn struct {
	ID          string  `json:"id" jsonschema:"要更新的项目 ID"`
	Title       *string `json:"title,omitempty" jsonschema:"新标题"`
	CategoryID  *string `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Role        *string `json:"role,omitempty" jsonschema:"新角色"`
	Client      *string `json:"client,omitempty" jsonschema:"新客户"`
	CoverURL    *string `json:"cover_url,omitempty" jsonschema:"新封面图 URL"`
	Summary     *string `json:"summary,omitempty" jsonschema:"新简介"`
	Description *string `json:"description,omitempty" jsonschema:"新详述（Markdown）"`
	TechStack   *string `json:"tech_stack,omitempty" jsonschema:"新技能标签，逗号分隔"`
	StartDate   *string `json:"start_date,omitempty" jsonschema:"新开始月份 YYYY-MM；空串=清空"`
	EndDate     *string `json:"end_date,omitempty" jsonschema:"新结束月份 YYYY-MM；空串=清空（至今）"`
	ProjectURL  *string `json:"project_url,omitempty" jsonschema:"新项目地址"`
	RepoURL     *string `json:"repo_url,omitempty" jsonschema:"新仓库地址"`
	Status      *int    `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder   *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateProject 更新项目。
func (s *Service) updateProject(ctx context.Context, _ *sdk.CallToolRequest, in UpdateProjectIn) (*sdk.CallToolResult, res.ProjectRes, error) {
	if in.ID == "" {
		return nil, res.ProjectRes{}, fmt.Errorf("id 不能为空")
	}
	if in.StartDate != nil && *in.StartDate != "" {
		if err := validateMonth(*in.StartDate, "start_date"); err != nil {
			return nil, res.ProjectRes{}, err
		}
	}
	if in.EndDate != nil && *in.EndDate != "" {
		if err := validateMonth(*in.EndDate, "end_date"); err != nil {
			return nil, res.ProjectRes{}, err
		}
	}
	updated, err := s.projectLogic.Update(ctx, in.ID, &req.UpdateProjectReq{
		Title:       in.Title,
		CategoryID:  in.CategoryID,
		Role:        in.Role,
		Client:      in.Client,
		CoverURL:    in.CoverURL,
		Summary:     in.Summary,
		Description: in.Description,
		TechStack:   in.TechStack,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		ProjectURL:  in.ProjectURL,
		RepoURL:     in.RepoURL,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.ProjectRes{}, fmt.Errorf("更新项目失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteProject 删除项目。
func (s *Service) deleteProject(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.projectLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除项目失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}

// validateMonth 校验 YYYY-MM 月份格式（与后端按月粒度的约定一致）。
func validateMonth(v, field string) error {
	if v == "" {
		return nil
	}
	if len(v) != 7 || v[4] != '-' {
		return fmt.Errorf("%s 格式应为 YYYY-MM，实际为 %q", field, v)
	}
	for _, r := range v[:4] + v[5:] {
		if r < '0' || r > '9' {
			return fmt.Errorf("%s 格式应为 YYYY-MM，实际为 %q", field, v)
		}
	}
	return nil
}
