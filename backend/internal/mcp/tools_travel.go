package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTravelTools 注册旅行攻略工具。
func (s *Service) registerTravelTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_travels",
		Description: "分页查询旅行攻略列表（不含景点/行程等 JSON 大字段），可按关键词、地区、状态、天数过滤。",
	}, wrapTool(s, "list_travels", s.listTravels))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_travel",
		Description: "按 ID 获取旅行攻略详情（含景点、行程、点评完整数据）。",
	}, wrapTool(s, "get_travel", s.getTravel))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_travel",
		Description: "新增旅行攻略。title/destination/region 必填；status: draft(草稿，默认)/published(发布)/offline(下架)；attractions/itinerary 为对象数组整体写入。",
	}, wrapTool(s, "create_travel", s.createTravel))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_travel",
		Description: "更新旅行攻略（整体替换传入了的字段，未传入的保持不变；attractions/itinerary/reviews 传入时整体替换）。",
	}, wrapTool(s, "update_travel", s.updateTravel))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "set_travel_status",
		Description: "修改旅行攻略状态。action 取值：draft（转草稿）、publish（发布）、offline（下架）。",
	}, wrapTool(s, "set_travel_status", s.setTravelStatus))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_travel",
		Description: "删除旅行攻略（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_travel", s.deleteTravel))
}

// ListTravelsIn list_travels 入参。
type ListTravelsIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题/目的地/简介模糊搜索"`
	Region     string `json:"region,omitempty" jsonschema:"按地区筛选"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=travel 的分类）"`
	Status     string `json:"status,omitempty" jsonschema:"按状态筛选: draft/published/offline，不传返回全部"`
	DaysRange  string `json:"days_range,omitempty" jsonschema:"按行程天数筛选: 1-3 / 4-7 / 8-14 / 15+"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListTravelsOut list_travels 出参。
type ListTravelsOut struct {
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Items []res.TravelGuideRes `json:"items"`
}

// listTravels 分页查询旅行攻略列表。
func (s *Service) listTravels(ctx context.Context, _ *sdk.CallToolRequest, in ListTravelsIn) (*sdk.CallToolResult, ListTravelsOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	var statusPtr *int16
	if in.Status != "" {
		status, err := statusFromString(in.Status)
		if err != nil {
			return nil, ListTravelsOut{}, err
		}
		if status != 0 {
			statusPtr = &status
		}
	}
	page, err := s.travelLogic.GetList(ctx, &req.TravelGuideListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		Region:     stringPtr(in.Region),
		CategoryID: stringPtr(in.CategoryID),
		Status:     statusPtr,
		DaysRange:  stringPtr(in.DaysRange),
	})
	if err != nil {
		return nil, ListTravelsOut{}, fmt.Errorf("查询旅行攻略列表失败: %w", err)
	}
	return nil, ListTravelsOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetTravelIn get_travel 入参。
type GetTravelIn struct {
	ID string `json:"id" jsonschema:"旅行攻略 ID"`
}

// getTravel 获取旅行攻略详情。
func (s *Service) getTravel(ctx context.Context, _ *sdk.CallToolRequest, in GetTravelIn) (*sdk.CallToolResult, res.TravelGuideDetailRes, error) {
	if in.ID == "" {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.travelLogic.GetDetail(ctx, in.ID)
	if err != nil {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("获取旅行攻略失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateTravelIn create_travel 入参。
type CreateTravelIn struct {
	Title       string           `json:"title" jsonschema:"攻略标题（必填，1-200 字）"`
	Destination string           `json:"destination" jsonschema:"目的地（必填，如 厦门）"`
	Region      string           `json:"region" jsonschema:"所属地区（必填，如 华东）"`
	Summary     string           `json:"summary,omitempty" jsonschema:"攻略简介"`
	CoverImage  string           `json:"cover_image,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Status      string           `json:"status,omitempty" jsonschema:"状态: draft(草稿，默认)/published(发布)/offline(下架)"`
	CategoryID  string           `json:"category_id,omitempty" jsonschema:"分类 ID（type=travel 的分类）"`
	Days        int              `json:"days,omitempty" jsonschema:"行程天数，默认 1"`
	BestMonth   string           `json:"best_month,omitempty" jsonschema:"最佳出行月份，如 5-6 月"`
	Attractions []map[string]any `json:"attractions,omitempty" jsonschema:"景点列表，对象数组整体写入，如 [{\"name\":\"鼓浪屿\",\"description\":\"...\"}]"`
	Itinerary   []map[string]any `json:"itinerary,omitempty" jsonschema:"每日行程，对象数组整体写入，如 [{\"day\":1,\"title\":\"...\",\"items\":[...]}]"`
	Reviews     []map[string]any `json:"reviews,omitempty" jsonschema:"自定义点评，对象数组整体写入"`
}

// createTravel 新增旅行攻略。
func (s *Service) createTravel(ctx context.Context, _ *sdk.CallToolRequest, in CreateTravelIn) (*sdk.CallToolResult, res.TravelGuideDetailRes, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Destination) == "" || strings.TrimSpace(in.Region) == "" {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("title、destination、region 为必填项")
	}
	status, err := statusFromString(in.Status)
	if err != nil {
		return nil, res.TravelGuideDetailRes{}, err
	}
	created, err := s.travelLogic.Create(ctx, &req.CreateTravelGuideReq{
		Title:       in.Title,
		Summary:     in.Summary,
		CoverImage:  in.CoverImage,
		Status:      status,
		Destination: in.Destination,
		Region:      in.Region,
		CategoryID:  stringPtr(in.CategoryID),
		Days:        in.Days,
		BestMonth:   in.BestMonth,
		Attractions: in.Attractions,
		Itinerary:   in.Itinerary,
		Reviews:     in.Reviews,
	})
	if err != nil {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("创建旅行攻略失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateTravelIn update_travel 入参（指针字段=不传不修改）。
type UpdateTravelIn struct {
	ID          string           `json:"id" jsonschema:"要更新的旅行攻略 ID"`
	Title       *string          `json:"title,omitempty" jsonschema:"新标题"`
	Summary     *string          `json:"summary,omitempty" jsonschema:"新简介"`
	CoverImage  *string          `json:"cover_image,omitempty" jsonschema:"新封面图 URL"`
	Status      *string          `json:"status,omitempty" jsonschema:"新状态: draft/published/offline"`
	Destination *string          `json:"destination,omitempty" jsonschema:"新目的地"`
	Region      *string          `json:"region,omitempty" jsonschema:"新地区"`
	CategoryID  *string          `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Days        *int             `json:"days,omitempty" jsonschema:"新行程天数"`
	BestMonth   *string          `json:"best_month,omitempty" jsonschema:"新最佳出行月份"`
	Attractions []map[string]any `json:"attractions,omitempty" jsonschema:"替换后的完整景点列表；不传该字段则不修改"`
	Itinerary   []map[string]any `json:"itinerary,omitempty" jsonschema:"替换后的完整每日行程；不传该字段则不修改"`
	Reviews     []map[string]any `json:"reviews,omitempty" jsonschema:"替换后的完整点评列表；不传该字段则不修改"`
}

// updateTravel 更新旅行攻略。
func (s *Service) updateTravel(ctx context.Context, _ *sdk.CallToolRequest, in UpdateTravelIn) (*sdk.CallToolResult, res.TravelGuideDetailRes, error) {
	if in.ID == "" {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("id 不能为空")
	}
	patch := req.UpdateTravelGuideReq{
		Title:       in.Title,
		Summary:     in.Summary,
		CoverImage:  in.CoverImage,
		Destination: in.Destination,
		Region:      in.Region,
		CategoryID:  in.CategoryID,
		Days:        in.Days,
		BestMonth:   in.BestMonth,
		Attractions: in.Attractions,
		Itinerary:   in.Itinerary,
		Reviews:     in.Reviews,
	}
	if in.Status != nil {
		status, err := statusFromString(*in.Status)
		if err != nil {
			return nil, res.TravelGuideDetailRes{}, err
		}
		if status != 0 {
			patch.Status = &status
		}
	}
	updated, err := s.travelLogic.Update(ctx, in.ID, &patch)
	if err != nil {
		return nil, res.TravelGuideDetailRes{}, fmt.Errorf("更新旅行攻略失败: %w", err)
	}
	return nil, *updated, nil
}

// SetTravelStatusIn set_travel_status 入参。
type SetTravelStatusIn struct {
	ID     string `json:"id" jsonschema:"旅行攻略 ID"`
	Action string `json:"action" jsonschema:"目标状态: draft(转草稿) / publish(发布) / offline(下架)"`
}

// setTravelStatus 修改旅行攻略状态。
func (s *Service) setTravelStatus(ctx context.Context, _ *sdk.CallToolRequest, in SetTravelStatusIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	status, err := statusFromString(in.Action)
	if err != nil {
		return nil, DeleteOut{}, err
	}
	if status == 0 {
		return nil, DeleteOut{}, fmt.Errorf("action 不能为空，可选：draft / publish / offline")
	}
	if err := s.travelLogic.UpdateStatus(ctx, in.ID, &req.UpdateTravelGuideStatusReq{Status: status}); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("更新旅行攻略状态失败: %w", err)
	}
	return nil, DeleteOut{Deleted: false, ID: in.ID}, nil
}

// deleteTravel 删除旅行攻略。
func (s *Service) deleteTravel(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.travelLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除旅行攻略失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
