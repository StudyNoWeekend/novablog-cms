package mcp

import (
	"context"
	"fmt"
	"strings"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerGameTools 注册游戏库工具。
func (s *Service) registerGameTools(server *sdk.Server) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_games",
		Description: "分页查询游戏库列表（不含短评正文），可按关键词、平台、类型、游玩状态过滤。",
	}, wrapTool(s, "list_games", s.listGames))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_game",
		Description: "按 ID 获取游戏详情（含完整短评）。",
	}, wrapTool(s, "get_game", s.getGame))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_game",
		Description: "新增游戏到游戏库。play_status: want(想玩)/playing(在玩)/played(已玩)；rating 0-10。",
	}, wrapTool(s, "create_game", s.createGame))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_game",
		Description: "更新游戏信息（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_game", s.updateGame))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_game",
		Description: "从游戏库删除游戏（软删除）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_game", s.deleteGame))
}

// ListGamesIn list_games 入参。
type ListGamesIn struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"按标题模糊搜索"`
	Platform   string `json:"platform,omitempty" jsonschema:"按平台模糊筛选"`
	Genre      string `json:"genre,omitempty" jsonschema:"按类型（genre）模糊筛选"`
	PlayStatus string `json:"play_status,omitempty" jsonschema:"按游玩状态筛选: want/playing/played，不传返回全部"`
	CategoryID string `json:"category_id,omitempty" jsonschema:"按分类 ID 筛选（type=game 的分类）"`
	Status     *int   `json:"status,omitempty" jsonschema:"按状态筛选: 0=草稿 1=已发布"`
	Page       int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListGamesOut list_games 出参。
type ListGamesOut struct {
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Items []res.GameCardRes `json:"items"`
}

// listGames 分页查询游戏列表。
func (s *Service) listGames(ctx context.Context, _ *sdk.CallToolRequest, in ListGamesIn) (*sdk.CallToolResult, ListGamesOut, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.gameLogic.GetList(ctx, &req.GameListReq{
		PageReq:    req.PageReq{Page: in.Page, PageSize: in.PageSize},
		Keyword:    stringPtr(in.Keyword),
		Platform:   stringPtr(in.Platform),
		Genre:      stringPtr(in.Genre),
		PlayStatus: stringPtr(in.PlayStatus),
		CategoryID: stringPtr(in.CategoryID),
		Status:     in.Status,
	})
	if err != nil {
		return nil, ListGamesOut{}, fmt.Errorf("查询游戏列表失败: %w", err)
	}
	return nil, ListGamesOut{Total: page.Total, Page: page.Page, Items: page.List}, nil
}

// GetGameIn get_game 入参。
type GetGameIn struct {
	ID string `json:"id" jsonschema:"游戏 ID"`
}

// getGame 获取游戏详情。
func (s *Service) getGame(ctx context.Context, _ *sdk.CallToolRequest, in GetGameIn) (*sdk.CallToolResult, res.GameRes, error) {
	if in.ID == "" {
		return nil, res.GameRes{}, fmt.Errorf("id 不能为空")
	}
	detail, err := s.gameLogic.GetByID(ctx, in.ID)
	if err != nil {
		return nil, res.GameRes{}, fmt.Errorf("获取游戏失败: %w", err)
	}
	return nil, *detail, nil
}

// CreateGameIn create_game 入参。
type CreateGameIn struct {
	Title       string `json:"title" jsonschema:"游戏名（必填，1-255 字）"`
	Cover       string `json:"cover,omitempty" jsonschema:"封面图 URL，可先调用 upload_image 上传获取"`
	Platform    string `json:"platform,omitempty" jsonschema:"平台（如 PC/Switch/PS5）"`
	Genre       string `json:"genre,omitempty" jsonschema:"类型（如 RPG/独立游戏）"`
	PlayStatus  string `json:"play_status,omitempty" jsonschema:"游玩状态: want(想玩)/playing(在玩)/played(已玩)，默认 want"`
	PlayHours   *int   `json:"play_hours,omitempty" jsonschema:"游戏时长（小时，≥0）"`
	Rating      *int   `json:"rating,omitempty" jsonschema:"评分 0-10"`
	ShortReview string `json:"short_review,omitempty" jsonschema:"短评（≤1000 字）"`
	CategoryID  string `json:"category_id,omitempty" jsonschema:"分类 ID（type=game 的分类）"`
	Status      *int   `json:"status,omitempty" jsonschema:"状态: 0=草稿(默认) 1=已发布"`
	SortOrder   *int   `json:"sort_order,omitempty" jsonschema:"排序权重"`
}

// createGame 新增游戏。
func (s *Service) createGame(ctx context.Context, _ *sdk.CallToolRequest, in CreateGameIn) (*sdk.CallToolResult, res.GameRes, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, res.GameRes{}, fmt.Errorf("title 不能为空")
	}
	if in.Rating != nil && (*in.Rating < 0 || *in.Rating > 10) {
		return nil, res.GameRes{}, fmt.Errorf("rating 取值 0-10")
	}
	switch in.PlayStatus {
	case "", "want", "playing", "played":
	default:
		return nil, res.GameRes{}, fmt.Errorf("play_status 取值非法：want / playing / played")
	}
	created, err := s.gameLogic.Create(ctx, &req.CreateGameReq{
		Title:       in.Title,
		Cover:       in.Cover,
		Platform:    in.Platform,
		Genre:       in.Genre,
		PlayStatus:  in.PlayStatus,
		PlayHours:   in.PlayHours,
		Rating:      in.Rating,
		ShortReview: in.ShortReview,
		CategoryID:  stringPtr(in.CategoryID),
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.GameRes{}, fmt.Errorf("创建游戏失败: %w", err)
	}
	return nil, *created, nil
}

// UpdateGameIn update_game 入参（指针字段=不传不修改）。
type UpdateGameIn struct {
	ID          string  `json:"id" jsonschema:"要更新的游戏 ID"`
	Title       *string `json:"title,omitempty" jsonschema:"新游戏名"`
	Cover       *string `json:"cover,omitempty" jsonschema:"新封面图 URL"`
	Platform    *string `json:"platform,omitempty" jsonschema:"新平台"`
	Genre       *string `json:"genre,omitempty" jsonschema:"新类型"`
	PlayStatus  *string `json:"play_status,omitempty" jsonschema:"新游玩状态: want/playing/played"`
	PlayHours   *int    `json:"play_hours,omitempty" jsonschema:"新游戏时长（小时）"`
	Rating      *int    `json:"rating,omitempty" jsonschema:"新评分 0-10"`
	ShortReview *string `json:"short_review,omitempty" jsonschema:"新短评"`
	CategoryID  *string `json:"category_id,omitempty" jsonschema:"新分类 ID"`
	Status      *int    `json:"status,omitempty" jsonschema:"新状态: 0=草稿 1=已发布"`
	SortOrder   *int    `json:"sort_order,omitempty" jsonschema:"新排序权重"`
}

// updateGame 更新游戏。
func (s *Service) updateGame(ctx context.Context, _ *sdk.CallToolRequest, in UpdateGameIn) (*sdk.CallToolResult, res.GameRes, error) {
	if in.ID == "" {
		return nil, res.GameRes{}, fmt.Errorf("id 不能为空")
	}
	if in.Rating != nil && (*in.Rating < 0 || *in.Rating > 10) {
		return nil, res.GameRes{}, fmt.Errorf("rating 取值 0-10")
	}
	if in.PlayStatus != nil {
		switch *in.PlayStatus {
		case "want", "playing", "played":
		default:
			return nil, res.GameRes{}, fmt.Errorf("play_status 取值非法：want / playing / played")
		}
	}
	updated, err := s.gameLogic.Update(ctx, in.ID, &req.UpdateGameReq{
		Title:       in.Title,
		Cover:       in.Cover,
		Platform:    in.Platform,
		Genre:       in.Genre,
		PlayStatus:  in.PlayStatus,
		PlayHours:   in.PlayHours,
		Rating:      in.Rating,
		ShortReview: in.ShortReview,
		CategoryID:  in.CategoryID,
		Status:      in.Status,
		SortOrder:   in.SortOrder,
	})
	if err != nil {
		return nil, res.GameRes{}, fmt.Errorf("更新游戏失败: %w", err)
	}
	return nil, *updated, nil
}

// deleteGame 删除游戏。
func (s *Service) deleteGame(ctx context.Context, _ *sdk.CallToolRequest, in DeleteIn) (*sdk.CallToolResult, DeleteOut, error) {
	if in.ID == "" {
		return nil, DeleteOut{}, fmt.Errorf("id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteOut{}, fmt.Errorf("删除是不可逆操作，请确认后显式传 confirm=true 再执行")
	}
	if err := s.gameLogic.Delete(ctx, in.ID); err != nil {
		return nil, DeleteOut{}, fmt.Errorf("删除游戏失败: %w", err)
	}
	return nil, DeleteOut{Deleted: true, ID: in.ID}, nil
}
