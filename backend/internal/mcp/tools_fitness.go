package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerFitnessTools 注册健身训练工具。
func (s *Service) registerFitnessTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_fitness_records",
		Description: "分页查询健身训练记录列表（不含动作清单明细），按训练日期倒序。",
	}, wrapTool(s, "list_fitness_records", s.listFitnessRecords))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_fitness_record",
		Description: "按 ID 获取训练记录详情（含动作清单）。",
	}, wrapTool(s, "get_fitness_record", s.getFitnessRecord))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_fitness_record",
		Description: "新增一条健身训练记录。date 不传默认今天；type: strength(力量)/cardio(有氧)/stretch(拉伸)；content 为动作清单对象数组。",
	}, wrapTool(s, "create_fitness_record", s.createFitnessRecord))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_fitness_record",
		Description: "更新训练记录（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_fitness_record", s.updateFitnessRecord))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_fitness_record",
		Description: "删除训练记录（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_fitness_record", s.deleteFitnessRecord))
}

// ListFitnessRecordsIn list_fitness_records 入参。
type ListFitnessRecordsIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=fitness 的分类）"`
	Type       string `json:"type,omitempty" jsonschema:"按训练类型筛选: strength/cardio/stretch，不传返回全部"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListFitnessRecordsOut list_fitness_records 出参。
type ListFitnessRecordsOut struct {
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Items []res.FitnessCardRes `json:"items"`
}

// listFitnessRecords 分页查询训练记录列表。
func (s *Service) listFitnessRecords(ctx context.Context, _ *sdk.CallToolRequest, in ListFitnessRecordsIn) (*sdk.CallToolResult, ListFitnessRecordsOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.fitnessLogic.GetList(ctx, &req.FitnessListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Type:       stringPtr(in.Type),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListFitnessRecordsOut{}, fmt.Errorf("查询训练记录列表失败: %w", err)
	}
	return nil, ListFitnessRecordsOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetFitnessRecordIn get_fitness_record 入参。
type GetFitnessRecordIn struct {
	ID string `json:"id" jsonschema:"训练记录 ID"`
}

// getFitnessRecord 获取训练记录详情。
func (s *Service) getFitnessRecord(ctx context.Context, _ *sdk.CallToolRequest, in GetFitnessRecordIn) (*sdk.CallToolResult, res.FitnessRes, error) {
	if in.ID == "" {
		return nil, res.FitnessRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.fitnessLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.FitnessRes{}, fmt.Errorf("获取训练记录失败: %w", err)
	}
	return nil, *detail, nil
}

// FitnessExerciseIn 训练动作入参。
type FitnessExerciseIn struct {
	Name string `json:"name" jsonschema:"动作名称，如 深蹲"`
	Sets int    `json:"sets,omitempty" jsonschema:"组数"`
	Reps string `json:"reps,omitempty" jsonschema:"次数描述，如 12次 或 8-10次"`
	Note string `json:"note,omitempty" jsonschema:"备注"`
}

// CreateFitnessRecordIn create_fitness_record 入参。
type CreateFitnessRecordIn struct {
	Date        *time.Time          `json:"date,omitempty" jsonschema:"训练日期（RFC3339，不传默认今天）"`
	Title       string              `json:"title" jsonschema:"训练标题（必填，1-255 字，如 腿日力量训练）"`
	CategoryID  string              `json:"category_id,omitempty" jsonschema:"分类 ID（type=fitness 的分类，设置后 type 镜像为分类 slug）"`
	Type        string              `json:"type,omitempty" jsonschema:"训练类型: strength(力量)/cardio(有氧)/stretch(拉伸)；建议改用 category_id"`
	DurationMin *int                `json:"duration_min,omitempty" jsonschema:"训练时长（分钟，≥0）"`
	Calories    *int                `json:"calories,omitempty" jsonschema:"消耗热量（千卡，≥0）"`
	Content     []FitnessExerciseIn `json:"content,omitempty" jsonschema:"动作清单，如 [{\"name\":\"深蹲\",\"sets\":4,\"reps\":\"12次\",\"note\":\"...\"}]"`
	Notes       string              `json:"notes,omitempty" jsonschema:"训练备注（Markdown）"`
	Status      *int                `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int                `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createFitnessRecord 新增训练记录。
func (s *Service) createFitnessRecord(ctx context.Context, _ *sdk.CallToolRequest, in CreateFitnessRecordIn) (*sdk.CallToolResult, res.FitnessRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.FitnessRes{}, fmt.Errorf("title 不能为空")
	}
	if in.Type != "" {
		switch in.Type {
		case "strength", "cardio", "stretch":
		default:
			return nil, res.FitnessRes{}, fmt.Errorf("type 取值非法：strength / cardio / stretch")
		}
	}
	content := make([]map[string]any, 0, len(in.Content))
	for _, c := range in.Content {
		if strings.TrimSpace(c.Name) == "" {
			return nil, res.FitnessRes{}, fmt.Errorf("content 中每个动作必须提供 name")
		}
		entry := map[string]any{"name": c.Name}
		if c.Sets > 0 {
			entry["sets"] = c.Sets
		}
		if c.Reps != "" {
			entry["reps"] = c.Reps
		}
		if c.Note != "" {
			entry["note"] = c.Note
		}
		content = append(content, entry)
	}
	created, err := s.fitnessLogic.Create(ctx, &req.CreateFitnessReq{
		Date:        in.Date,
		Title:       in.Title,
		CategoryID:  stringPtr(in.CategoryID),
		Type:        in.Type,
		DurationMin: in.DurationMin,
		Calories:    in.Calories,
		Content:     content,
		Notes:       in.Notes,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.FitnessRes{}, fmt.Errorf("创建训练记录失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateFitnessRecordIn update_fitness_record 入参（指针字段=不传不修改）。
type UpdateFitnessRecordIn struct {
	ID          string              `json:"id" jsonschema:"要更新的训练记录 ID"`
	Date        *time.Time          `json:"date,omitempty" jsonschema:"新训练日期（RFC3339）"`
	Title       *string             `json:"title,omitempty" jsonschema:"新标题"`
	CategoryID  *string             `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Type        *string             `json:"type,omitempty" jsonschema:"新训练类型: strength/cardio/stretch"`
	DurationMin *int                `json:"duration_min,omitempty" jsonschema:"新训练时长（分钟）"`
	Calories    *int                `json:"calories,omitempty" jsonschema:"新消耗热量（千卡）"`
	Content     []FitnessExerciseIn `json:"content,omitempty" jsonschema:"替换后的完整动作清单；不传该字段则不修改"`
	Notes       *string             `json:"notes,omitempty" jsonschema:"新备注（Markdown）"`
	Status      *int                `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder   *int                `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateFitnessRecord 更新训练记录。
func (s *Service) updateFitnessRecord(ctx context.Context, _ *sdk.CallToolRequest, in UpdateFitnessRecordIn) (*sdk.CallToolResult, res.FitnessRes, error) {
	if in.ID == "" {
		return nil, res.FitnessRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Type != nil {
		switch *in.Type {
		case "strength", "cardio", "stretch":
		default:
			return nil, res.FitnessRes{}, fmt.Errorf("type 取值非法：strength / cardio / stretch")
		}
	}
	content := make([]map[string]any, 0, len(in.Content))
	for _, c := range in.Content {
		if strings.TrimSpace(c.Name) == "" {
			return nil, res.FitnessRes{}, fmt.Errorf("content 中每个动作必须提供 name")
		}
		entry := map[string]any{"name": c.Name}
		if c.Sets > 0 {
			entry["sets"] = c.Sets
		}
		if c.Reps != "" {
			entry["reps"] = c.Reps
		}
		if c.Note != "" {
			entry["note"] = c.Note
		}
		content = append(content, entry)
	}
	updated, err := s.fitnessLogic.Update(ctx, in.ID, &req.UpdateFitnessReq{
		Date:        in.Date,
		Title:       in.Title,
		CategoryID:  in.CategoryID,
		Type:        in.Type,
		DurationMin: in.DurationMin,
		Calories:    in.Calories,
		Content:     content,
		Notes:       in.Notes,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.FitnessRes{}, fmt.Errorf("更新训练记录失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteFitnessRecord 删除训练记录。
func (s *Service) deleteFitnessRecord(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.fitnessLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除训练记录失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
