package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerVideoTools 注册视频作品工具。
func (s *Service) registerVideoTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_videos",
		Description: "分页查询视频作品列表（含各平台链接），可按关键词、状态、分类过滤。",
	}, wrapTool(s, "list_videos", s.listVideos))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_video",
		Description: "按 ID 获取视频作品详情（含平台链接列表）。",
	}, wrapTool(s, "get_video", s.getVideo))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_video",
		Description: "新增视频作品。platforms 必填且至少一条（B站/YouTube 等，整体替换语义）。",
	}, wrapTool(s, "create_video", s.createVideo))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_video",
		Description: "更新视频作品（整体替换传入了的字段，未传入的保持不变；platforms 传入时整体替换全部平台链接，不传则不修改）。",
	}, wrapTool(s, "update_video", s.updateVideo))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_video",
		Description: "删除视频作品（级联软删平台链接）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_video", s.deleteVideo))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "parse_video_url",
		Description: "解析视频链接的元信息（标题/封面/描述，支持 B 站等平台），用于创建视频作品前自动填充字段。",
	}, wrapTool(s, "parse_video_url", s.parseVideoURL))
}

// ListVideosIn list_videos 入参。
type ListVideosIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题模糊搜索"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=video 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListVideosOut list_videos 出参。
type ListVideosOut struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Items []res.VideoWorkRes `json:"items"`
}

// listVideos 分页查询视频列表。
func (s *Service) listVideos(ctx context.Context, _ *sdk.CallToolRequest, in ListVideosIn) (*sdk.CallToolResult, ListVideosOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.videoLogic.GetList(ctx, &req.VideoListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListVideosOut{}, fmt.Errorf("查询视频列表失败: %w", err)
	}
	return nil, ListVideosOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetVideoIn get_video 入参。
type GetVideoIn struct {
	ID string `json:"id" jsonschema:"视频作品 ID"`
}

// getVideo 获取视频详情。
func (s *Service) getVideo(ctx context.Context, _ *sdk.CallToolRequest, in GetVideoIn) (*sdk.CallToolResult, res.VideoWorkRes, error) {
	if in.ID == "" {
		return nil, res.VideoWorkRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.videoLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.VideoWorkRes{}, fmt.Errorf("获取视频失败: %w", err)
	}
	return nil, *detail, nil
}

// VideoPlatformIn 视频平台链接入参。
type VideoPlatformIn struct {
	Platform string `json:"platform" jsonschema:"平台标识（如 bilibili/youtube）"`
	URL      string `json:"url" jsonschema:"视频在该平台的地址"`
}

// CreateVideoIn create_video 入参。
type CreateVideoIn struct {
	Title       string                `json:"title" jsonschema:"视频标题（必填，1-255 字）"`
	CoverURL    string                `json:"cover_url,omitempty" jsonschema:"封面图 URL，可先调用 parse_video_url 或 upload_image 获取"`
	Description string                `json:"description,omitempty" jsonschema:"视频描述"`
	CategoryID  string                `json:"category_id,omitempty" jsonschema:"分类 ID（type=video 的分类）"`
	Status      *int                  `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	Platforms   []req.PlatformLinkReq `json:"platforms" jsonschema:"发布的平台链接列表（至少 1 条），如 [{\"platform\":\"bilibili\",\"url\":\"https://...\"}]"`
}

// createVideo 新增视频作品。
func (s *Service) createVideo(ctx context.Context, _ *sdk.CallToolRequest, in CreateVideoIn) (*sdk.CallToolResult, res.VideoWorkRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.VideoWorkRes{}, fmt.Errorf("title 不能为空")
	}
	if len(in.Platforms) == 0 {
		return nil, res.VideoWorkRes{}, fmt.Errorf("platforms 不能为空，至少提供一条平台链接")
	}
	created, err := s.videoLogic.Create(ctx, &req.CreateVideoReq{
		Title:       in.Title,
		CoverURL:    in.CoverURL,
		Description: in.Description,
		CategoryID:  stringPtr(in.CategoryID),
		Status:      in.Status,
		Platforms:   in.Platforms,
	})
	if err != nil {
		return nil, res.VideoWorkRes{}, fmt.Errorf("创建视频失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateVideoIn update_video 入参（指针字段=不传不修改；platforms 传入时整体替换）。
type UpdateVideoIn struct {
	ID          string                `json:"id" jsonschema:"要更新的视频作品 ID"`
	Title       *string               `json:"title,omitempty" jsonschema:"新标题"`
	CoverURL    *string               `json:"cover_url,omitempty" jsonschema:"新封面图 URL"`
	Description *string               `json:"description,omitempty" jsonschema:"新描述"`
	CategoryID  *string               `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Status      *int                  `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	Platforms   []req.PlatformLinkReq `json:"platforms,omitempty" jsonschema:"替换后的完整平台链接列表；不传该字段则不修改，传空数组则清空"`
}

// updateVideo 更新视频作品。
func (s *Service) updateVideo(ctx context.Context, _ *sdk.CallToolRequest, in UpdateVideoIn) (*sdk.CallToolResult, res.VideoWorkRes, error) {
	if in.ID == "" {
		return nil, res.VideoWorkRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Platforms != nil {
		for _, p := range in.Platforms {
			if strings.TrimSpace(p.Platform) == "" || strings.TrimSpace(p.URL) == "" {
				return nil, res.VideoWorkRes{}, fmt.Errorf("platforms 中 platform 与 url 均为必填")
			}
		}
	}
	updated, err := s.videoLogic.Update(ctx, in.ID, &req.UpdateVideoReq{
		Title:       in.Title,
		CoverURL:    in.CoverURL,
		Description: in.Description,
		CategoryID:  in.CategoryID,
		Status:      in.Status,
		Platforms:   in.Platforms,
	})
	if err != nil {
		return nil, res.VideoWorkRes{}, fmt.Errorf("更新视频失败: %w", err)
	}
	return nil, *updated, nil
}

// ParseVideoURLIn parse_video_url 入参。
type ParseVideoURLIn struct {
	Platform string `json:"platform" jsonschema:"平台标识（如 bilibili）"`
	URL      string `json:"url" jsonschema:"视频页面地址"`
}

// ParseVideoURLOut parse_video_url 出参。
type ParseVideoURLOut struct {
	Title       string `json:"title" jsonschema:"解析出的视频标题"`
	CoverURL    string `json:"cover_url" jsonschema:"解析出的封面图 URL"`
	Description string `json:"description" jsonschema:"解析出的视频简介"`
}

// parseVideoURL 解析视频链接元信息。
func (s *Service) parseVideoURL(ctx context.Context, _ *sdk.CallToolRequest, in ParseVideoURLIn) (*sdk.CallToolResult, ParseVideoURLOut, error) {
	if strings.TrimSpace(in.Platform) == "" || strings.TrimSpace(in.URL) == "" {
		return nil, ParseVideoURLOut{}, fmt.Errorf("platform 与 url 为必填项")
	}
	parsed, err := s.videoLogic.ParseVideo(ctx, &req.ParseVideoReq{Platform: in.Platform, URL: in.URL})
	if err != nil {
		return nil, ParseVideoURLOut{}, fmt.Errorf("解析视频信息失败: %w", err)
	}
	return nil, ParseVideoURLOut{Title: parsed.Title, CoverURL: parsed.CoverURL, Description: parsed.Description}, nil
}

// DeleteVideoIn delete_video 入参。
type DeleteVideoIn struct {
	ID      string `json:"id" jsonschema:"要删除的视频作品 ID"`
	Confirm bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// DeleteOut delete_video 等删除操作的统一出参。
type DeleteVideoOut struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// deleteVideo 删除视频作品。
func (s *Service) deleteVideo(ctx context.Context, _ *sdk.CallToolRequest, in DeleteVideoIn) (*sdk.CallToolResult, DeleteVideoOut, error) {
	if in.ID == "" {
		return nil, DeleteVideoOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteVideoOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.videoLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteVideoOut{}, fmt.Errorf("删除视频失败: %w", err)
	}
	return nil, DeleteVideoOut{Deleted: true, ID: in.ID}, nil
}
