package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPlaylistTools 注册第三方歌单工具。
func (s *Service) registerPlaylistTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_playlists",
		Description: "分页查询第三方歌单列表（网易云/QQ 音乐等外链歌单，含完整字段）。",
	}, wrapTool(s, "list_playlists", s.listPlaylists))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_playlist",
		Description: "按 ID 获取第三方歌单详情。",
	}, wrapTool(s, "get_playlist", s.getPlaylist))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_playlist",
		Description: "新增第三方歌单。title/platform/platform_url 必填；enabled 默认 true。",
	}, wrapTool(s, "create_playlist", s.createPlaylist))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_playlist",
		Description: "更新第三方歌单（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_playlist", s.updatePlaylist))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_playlist",
		Description: "删除第三方歌单（硬删除，不可恢复）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_playlist", s.deletePlaylist))
}

// ListPlaylistsIn list_playlists 入参。
type ListPlaylistsIn struct {
	Page     int `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize int `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListPlaylistsOut list_playlists 出参。
type ListPlaylistsOut struct {
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Items []res.PlaylistRes `json:"items"`
}

// listPlaylists 分页查询歌单列表。
func (s *Service) listPlaylists(ctx context.Context, _ *sdk.CallToolRequest, in ListPlaylistsIn) (*sdk.CallToolResult, ListPlaylistsOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.playlistLogic.GetList(ctx, &req.PlaylistListReq{
		PageReq: req.PageReq{Page: in.Page, PageSize: in.PageSize},
	})
	if err != nil {
		return nil, ListPlaylistsOut{}, fmt.Errorf("查询歌单列表失败: %w", err)
	}
	return nil, ListPlaylistsOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetPlaylistIn get_playlist 入参。
type GetPlaylistIn struct {
	ID string `json:"id" jsonschema:"歌单 ID"`
}

// getPlaylist 获取歌单详情。
func (s *Service) getPlaylist(ctx context.Context, _ *sdk.CallToolRequest, in GetPlaylistIn) (*sdk.CallToolResult, res.PlaylistRes, error) {
	if in.ID == "" {
		return nil, res.PlaylistRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.playlistLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.PlaylistRes{}, fmt.Errorf("获取歌单失败: %w", err)
	}
	return nil, *detail, nil
}

// CreatePlaylistIn create_playlist 入参。
type CreatePlaylistIn struct {
	Title       string `json:"title" jsonschema:"歌单标题（必填，1-500 字）"`
	CoverURL    string `json:"cover_url,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Platform    string `json:"platform" jsonschema:"音乐平台（必填，如网易云/QQ音乐）"`
	PlatformURL string `json:"platform_url" jsonschema:"歌单外链地址（必填）"`
	Description string `json:"description,omitempty" jsonschema:"歌单描述"`
	SortOrder   int    `json:"sort_order,omitempty" jsonschema:"排序权重"`
	Enabled     *bool  `json:"enabled,omitempty" jsonschema:"是否启用，默认 true"`
}

// createPlaylist 新增歌单。
func (s *Service) createPlaylist(ctx context.Context, _ *sdk.CallToolRequest, in CreatePlaylistIn) (*sdk.CallToolResult, res.PlaylistRes, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Platform) == "" || strings.TrimSpace(in.PlatformURL) == "" {
		return nil, res.PlaylistRes{}, fmt.Errorf("title、platform、platform_url 为必填项")
	}
	created, err := s.playlistLogic.Create(ctx, &req.CreatePlaylistReq{
		Title:       in.Title,
		CoverURL:    in.CoverURL,
		Platform:    in.Platform,
		PlatformURL: in.PlatformURL,
		Description: in.Description,
		SortOrder:   in.SortOrder,
		Enabled:     in.Enabled,
	})
	if err != nil {
		return nil, res.PlaylistRes{}, fmt.Errorf("创建歌单失败: %w", err)
	}
	return nil, *created, nil
}

// UpdatePlaylistIn update_playlist 入参（指针字段=不传不修改）。
type UpdatePlaylistIn struct {
	ID          string  `json:"id" jsonschema:"要更新的歌单 ID"`
	Title       *string `json:"title,omitempty" jsonschema:"新标题"`
	CoverURL    *string `json:"cover_url,omitempty" jsonschema:"新封面图 URL"`
	Platform    *string `json:"platform,omitempty" jsonschema:"新平台"`
	PlatformURL *string `json:"platform_url,omitempty" jsonschema:"新歌单外链"`
	Description *string `json:"description,omitempty" jsonschema:"新描述"`
	SortOrder   *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
	Enabled     *bool   `json:"enabled,omitempty" jsonschema:"是否启用"`
}

// updatePlaylist 更新歌单。
func (s *Service) updatePlaylist(ctx context.Context, _ *sdk.CallToolRequest, in UpdatePlaylistIn) (*sdk.CallToolResult, res.PlaylistRes, error) {
	if in.ID == "" {
		return nil, res.PlaylistRes{}, fmt.Errorf("id 不能为空")
	}
	updated, err := s.playlistLogic.Update(ctx, in.ID, &req.UpdatePlaylistReq{
		Title:       in.Title,
		CoverURL:    in.CoverURL,
		Platform:    in.Platform,
		PlatformURL: in.PlatformURL,
		Description: in.Description,
		SortOrder:   in.SortOrder,
		Enabled:     in.Enabled,
	})
	if err != nil {
		return nil, res.PlaylistRes{}, fmt.Errorf("更新歌单失败: %w", err)
	}
	return nil, *updated, nil
}

// deletePlaylist 删除歌单（硬删除）。
func (s *Service) deletePlaylist(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作（硬删除），请确认后显式传 confirm=true 再执行")
	}
	if err := s.playlistLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除歌单失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
