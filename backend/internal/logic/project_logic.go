package logic

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	"github.com/google/uuid"
)

// monthLayout 项目经历时间区间的输入输出格式（按月粒度）。
const monthLayout = "2006-01"

// ProjectLogic 项目经历业务逻辑结构体。
type ProjectLogic struct {
	projectModel *model.ProjectModel
}

// NewProjectLogic 创建 ProjectLogic 实例。
func NewProjectLogic() *ProjectLogic {
	return &ProjectLogic{
		projectModel: model.NewProject(),
	}
}

// Create 创建项目经历。
func (l *ProjectLogic) Create(ctx context.Context, r *req.CreateProjectReq) (*res.ProjectRes, error) {
	startDate, err := parseMonthPtr(r.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseMonthPtr(r.EndDate)
	if err != nil {
		return nil, err
	}

	status := 0 // 默认草稿
	if r.Status != nil {
		status = *r.Status
	}
	sortOrder := 0
	if r.SortOrder != nil {
		sortOrder = *r.SortOrder
	}

	p := &model.Project{
		ID:          uuid.New().String(),
		Title:       r.Title,
		Category:    r.Category,
		Role:        r.Role,
		Client:      r.Client,
		CoverURL:    r.CoverURL,
		Summary:     r.Summary,
		Description: r.Description,
		TechStack:   r.TechStack,
		StartDate:   startDate,
		EndDate:     endDate,
		ProjectURL:  r.ProjectURL,
		RepoURL:     r.RepoURL,
		Status:      status,
		SortOrder:   sortOrder,
	}
	if err := validateProjectRange(p); err != nil {
		return nil, err
	}

	if err := l.projectModel.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("创建项目经历失败: %w", err)
	}

	return l.toProjectRes(p), nil
}

// GetList 分页查询项目经历列表。
func (l *ProjectLogic) GetList(ctx context.Context, r *req.ProjectListReq) (*res.ProjectListRes, error) {
	list, total, err := l.projectModel.GetList(ctx, r.Keyword, r.Category, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询项目经历列表失败: %w", err)
	}

	items := make([]res.ProjectRes, 0, len(list))
	for i := range list {
		items = append(items, *l.toProjectRes(&list[i]))
	}

	return &res.ProjectListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询项目经历详情。
func (l *ProjectLogic) GetByID(ctx context.Context, id string) (*res.ProjectRes, error) {
	p, err := l.projectModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("项目经历不存在")
	}

	return l.toProjectRes(p), nil
}

// Update 更新项目经历（部分更新语义）。
func (l *ProjectLogic) Update(ctx context.Context, id string, r *req.UpdateProjectReq) (*res.ProjectRes, error) {
	p, err := l.projectModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("项目经历不存在")
	}

	if r.Title != nil {
		p.Title = *r.Title
	}
	if r.Category != nil {
		p.Category = *r.Category
	}
	if r.Role != nil {
		p.Role = *r.Role
	}
	if r.Client != nil {
		p.Client = *r.Client
	}
	if r.CoverURL != nil {
		p.CoverURL = *r.CoverURL
	}
	if r.Summary != nil {
		p.Summary = *r.Summary
	}
	if r.Description != nil {
		p.Description = *r.Description
	}
	if r.TechStack != nil {
		p.TechStack = *r.TechStack
	}
	if r.StartDate != nil {
		startDate, parseErr := parseMonthPtr(*r.StartDate)
		if parseErr != nil {
			return nil, parseErr
		}
		p.StartDate = startDate
	}
	if r.EndDate != nil {
		endDate, parseErr := parseMonthPtr(*r.EndDate)
		if parseErr != nil {
			return nil, parseErr
		}
		p.EndDate = endDate
	}
	if r.ProjectURL != nil {
		p.ProjectURL = *r.ProjectURL
	}
	if r.RepoURL != nil {
		p.RepoURL = *r.RepoURL
	}
	if r.Status != nil {
		p.Status = *r.Status
	}
	if r.SortOrder != nil {
		p.SortOrder = *r.SortOrder
	}
	if err := validateProjectRange(p); err != nil {
		return nil, err
	}

	if err := l.projectModel.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("更新项目经历失败: %w", err)
	}

	return l.toProjectRes(p), nil
}

// Delete 删除项目经历。
func (l *ProjectLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.projectModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("项目经历不存在")
	}

	if err := l.projectModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("软删除项目经历失败: %w", err)
	}

	return nil
}

// GetPublicList 获取已发布项目经历列表（强制 status=1）。
func (l *ProjectLogic) GetPublicList(ctx context.Context, r *req.ProjectListReq) (*res.ProjectListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布项目经历详情（未发布视为不存在）。
func (l *ProjectLogic) GetPublicDetail(ctx context.Context, id string) (*res.ProjectRes, error) {
	p, err := l.projectModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("项目经历不存在")
	}
	if p.Status != 1 {
		return nil, fmt.Errorf("项目经历不存在")
	}

	return l.toProjectRes(p), nil
}

// toProjectRes 转换为项目经历响应。
func (l *ProjectLogic) toProjectRes(p *model.Project) *res.ProjectRes {
	return &res.ProjectRes{
		ID:          p.ID,
		Title:       p.Title,
		Category:    p.Category,
		Role:        p.Role,
		Client:      p.Client,
		CoverURL:    p.CoverURL,
		Summary:     p.Summary,
		Description: p.Description,
		TechStack:   p.TechStack,
		StartDate:   formatMonthPtr(p.StartDate),
		EndDate:     formatMonthPtr(p.EndDate),
		ProjectURL:  p.ProjectURL,
		RepoURL:     p.RepoURL,
		Status:      p.Status,
		SortOrder:   p.SortOrder,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// parseMonthPtr 解析 YYYY-MM 字符串为当月 1 号，空字符串返回 nil（表示未设置/至今）。
func parseMonthPtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(monthLayout, s)
	if err != nil {
		return nil, fmt.Errorf("时间格式不正确，应为 YYYY-MM")
	}
	return &t, nil
}

// formatMonthPtr 将时间格式化为 YYYY-MM，nil 返回空字符串。
func formatMonthPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(monthLayout)
}

// validateProjectRange 校验时间区间：开始时间不能晚于结束时间。
func validateProjectRange(p *model.Project) error {
	if p.StartDate != nil && p.EndDate != nil && p.StartDate.After(*p.EndDate) {
		return fmt.Errorf("开始时间不能晚于结束时间")
	}
	return nil
}
