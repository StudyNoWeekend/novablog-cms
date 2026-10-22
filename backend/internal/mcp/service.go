// Package mcp 实现 NovaBlog 的 MCP（Model Context Protocol）服务端，
// 通过 Streamable HTTP 传输对外暴露内容管理工具，
// 供 Codex / ZCode 等 AI 客户端远程操作博客的全部内容模块。
package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"novablog/internal/logic"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// ctxKeyMCPKeyID Context 中 MCP 密钥 ID 的键（由 MCP 鉴权中间件注入请求上下文）。
type ctxKeyMCPKeyID struct{}

// WithKeyID 将 MCP 密钥 ID 注入请求上下文（由鉴权中间件调用）。
func WithKeyID(ctx context.Context, keyID string) context.Context {
	return context.WithValue(ctx, ctxKeyMCPKeyID{}, keyID)
}

// KeyIDFromContext 从请求上下文读取当前调用的 MCP 密钥 ID。
func KeyIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyMCPKeyID{}).(string); ok {
		return v
	}
	return ""
}

// Service MCP 服务端结构体：聚合各内容模块的业务 logic。
type Service struct {
	version string
	logger  *zap.Logger

	// 文章与媒体
	articleLogic     *logic.ArticleLogic
	mediaLogic       *logic.MediaLogic
	mediaFolderLogic *logic.MediaFolderLogic
	categoryLogic    *logic.CategoryLogic
	tagLogic         *logic.TagLogic

	// 关于我
	bloggerLogic *logic.BloggerLogic

	// 评论与统计
	commentLogic   *logic.CommentLogic
	analyticsLogic *logic.AnalyticsLogic
	viewStatsLogic *logic.ViewStatsLogic

	// 展示类内容模块
	equipmentLogic  *logic.EquipmentLogic
	projectLogic    *logic.ProjectLogic
	techStackLogic  *logic.TechStackLogic
	openSourceLogic *logic.OpenSourceLogic
	playlistLogic   *logic.ThirdPartyPlaylistLogic
	videoLogic      *logic.VideoLogic
	travelLogic     *logic.TravelGuideLogic
	portfolioLogic  *logic.PortfolioLogic
	recipeLogic     *logic.RecipeLogic
	bookLogic       *logic.BookLogic
	gameLogic       *logic.GameLogic
	fitnessLogic    *logic.FitnessLogic
}

// NewService 创建 MCP 服务端实例。
func NewService(version string, mediaLogic *logic.MediaLogic, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		version:          version,
		logger:           logger,
		articleLogic:     logic.NewArticleLogic(),
		mediaLogic:       mediaLogic,
		categoryLogic:    logic.NewCategoryLogic(),
		tagLogic:         logic.NewTagLogic(),
		bloggerLogic:     logic.NewBloggerLogic(),
		commentLogic:     logic.NewCommentLogic(),
		analyticsLogic:   logic.NewAnalyticsLogic(),
		viewStatsLogic:   logic.NewViewStatsLogic(),
		mediaFolderLogic: logic.NewMediaFolderLogic(),
		equipmentLogic:   logic.NewEquipmentLogic(),
		projectLogic:     logic.NewProjectLogic(),
		techStackLogic:   logic.NewTechStackLogic(),
		openSourceLogic:  logic.NewOpenSourceLogic(),
		playlistLogic:    logic.NewThirdPartyPlaylistLogic(),
		videoLogic:       logic.NewVideoLogic(),
		travelLogic:      logic.NewTravelGuideLogic(),
		portfolioLogic:   logic.NewPortfolioLogic(),
		recipeLogic:      logic.NewRecipeLogic(),
		bookLogic:        logic.NewBookLogic(),
		gameLogic:        logic.NewGameLogic(),
		fitnessLogic:     logic.NewFitnessLogic(),
	}
}

// Handler 返回 MCP Streamable HTTP 处理器（无状态模式，每个 POST 请求独立处理）。
func (s *Service) Handler() http.Handler {
	server := sdk.NewServer(&sdk.Implementation{
		Name:    "novablog",
		Version: s.version,
	}, &sdk.ServerOptions{
		Instructions: "NovaBlog 博客内容管理服务，覆盖管理后台全部内容模块。" +
			"可用能力：创建/更新/发布/下架/删除文章（Markdown），上传图片到媒体库，管理分类与标签，" +
			"管理关于我（博主资料、设备、项目经历、技术栈、开源作品），" +
			"管理内容创作模块（第三方歌单、视频作品、旅行攻略、摄影作品集、美食菜谱、读书书架、游戏库、健身训练），" +
			"查看工作台统计数据，管理评论（列表/回复/删除），查询/移动/删除媒体文件。" +
			"建议流程：先上传封面图与正文图片拿到 URL，再创建内容（默认保存为草稿，明确要求发布时才置为 published）；" +
			"删除操作均需显式传 confirm=true；复用已有分类与标签（先 list_categories / list_tags 查询）。",
	})
	s.registerTools(server)
	return sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server {
		return server
	}, &sdk.StreamableHTTPOptions{
		Stateless: true,
		// 关闭 SDK 的 DNS 重绑定防护：其判定条件是「连接本地地址为回环 + Host 头非回环」。
		// 本服务在单容器部署下由容器内 nginx 经 127.0.0.1 反代到 Go，Host 为公网域名，
		// 必然被误拦（403 invalid Host header）。该防护针对的是浏览器借 Cookie 等环境
		// 凭据攻击本地 MCP 服务的场景；本端点每次请求都需携带 Bearer Key（无环境凭据），
		// 且有外层 nginx 隔离与 IP 黑名单，防护所针对的风险不存在。
		DisableLocalhostProtection: true,
	})
}

// registerTools 注册全部 MCP 工具。
func (s *Service) registerTools(server *sdk.Server) {
	// 文章管理
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_articles",
		Description: "分页查询文章列表（含草稿），用于查找已有文章或确认标题是否重复。返回文章摘要信息，不含正文。",
	}, wrapTool(s, "list_articles", s.listArticles))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_article",
		Description: "按 ID 或 slug 获取文章完整内容（含 Markdown 正文），用于编辑前读取原文。",
	}, wrapTool(s, "get_article", s.getArticle))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_article",
		Description: "创建文章（Markdown 正文）。默认保存为草稿，status 传 published 才会直接发布。分类/标签按名称匹配，不存在时自动创建。",
	}, wrapTool(s, "create_article", s.createArticle))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_article",
		Description: "更新文章的字段（整体替换传入了的字段，未传入的保持不变）。按 ID 或 slug 定位。",
	}, wrapTool(s, "update_article", s.updateArticle))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "set_article_status",
		Description: "修改文章状态。action 取值：draft（转草稿）、publish（发布）、offline（下架）。",
	}, wrapTool(s, "set_article_status", s.setArticleStatus))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_article",
		Description: "删除文章（软删除）。必须显式传 confirm=true 才会执行，防止误删。",
	}, wrapTool(s, "delete_article", s.deleteArticle))

	// 媒体上传
	sdk.AddTool(server, &sdk.Tool{
		Name:        "upload_image",
		Description: "上传图片到媒体库并返回可直接引用的 URL。source 支持 http(s) 图片链接（自动转存）或 base64 图片数据。module 决定归入哪个模块文件夹（默认 article）。",
	}, wrapTool(s, "upload_image", s.uploadImage))

	// 媒体库管理
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_media",
		Description: "分页查询媒体库文件列表（含各模块文件夹归属信息），可按类型、关键词、文件夹过滤。",
	}, wrapTool(s, "list_media", s.listMedia))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_media_folders",
		Description: "查询媒体库文件夹列表（按内容模块组织的层级目录，已拍平并附完整路径 path 与文件数），用于选择 folder_id。",
	}, wrapTool(s, "list_media_folders", s.listMediaFolders))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "move_media",
		Description: "批量移动媒体文件到指定文件夹（folder_id 为空表示移到根目录）。",
	}, wrapTool(s, "move_media", s.moveMedia))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_media",
		Description: "删除媒体文件（含预设产物，硬删除）。文件被引用时需传 force=true；必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_media", s.deleteMedia))

	// 分类与标签
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_categories",
		Description: "查询分类列表（type 默认 article），创建内容前先查询以便复用。",
	}, wrapTool(s, "list_categories", s.listCategories))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_category",
		Description: "创建分类，type 为内容模块 key（如 article/travel/recipe 等）。",
	}, wrapTool(s, "create_category", s.createCategory))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_category",
		Description: "更新分类的名称/slug/描述/排序（整体替换传入了的字段，未传入的保持不变）。",
	}, wrapTool(s, "update_category", s.updateCategory))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_category",
		Description: "删除分类（被内容占用时会失败）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_category", s.deleteCategory))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_tags",
		Description: "查询标签列表，创建文章前先查询以便复用。",
	}, wrapTool(s, "list_tags", s.listTags))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_tag",
		Description: "创建标签。",
	}, wrapTool(s, "create_tag", s.createTag))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_tag",
		Description: "重命名标签。",
	}, wrapTool(s, "update_tag", s.updateTag))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_tag",
		Description: "删除标签。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_tag", s.deleteTag))

	// 博主资料（关于我）
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_profile",
		Description: "读取博主资料（关于我页面的全部信息：昵称、头像、简介、社交链接、博客标题等）。",
	}, wrapTool(s, "get_profile", s.getProfile))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_profile",
		Description: "更新博主资料（整体替换传入了的字段，未传入的保持不变；social_links/tags 传空数组表示清空）。",
	}, wrapTool(s, "update_profile", s.updateProfile))

	// 工作台统计（只读）
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_dashboard_overview",
		Description: "获取工作台概览：各内容模块的文章/发布/评论统计、流量摘要（今日/昨日 PV/UV、总 PV）与发布节奏。",
	}, wrapTool(s, "get_dashboard_overview", s.getDashboardOverview))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_dashboard_trend",
		Description: "获取内容产出趋势与流量趋势（range: 7d/30d/90d，默认 30d）。",
	}, wrapTool(s, "get_dashboard_trend", s.getDashboardTrend))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_top_content",
		Description: "获取热门内容排行（按浏览量或评论数）。type: article/travel/all；sort: views/comments。",
	}, wrapTool(s, "get_top_content", s.getTopContent))

	// 评论管理
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_comments",
		Description: "分页查询评论列表（含博主回复与访客评论），可按目标类型/目标 ID/关键词过滤。",
	}, wrapTool(s, "list_comments", s.listComments))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "reply_comment",
		Description: "以博主身份回复一条评论（1-2000 字）。",
	}, wrapTool(s, "reply_comment", s.replyComment))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_comment",
		Description: "删除评论（级联软删其下所有回复）。必须显式传 confirm=true 才会执行。",
	}, wrapTool(s, "delete_comment", s.deleteComment))

	// 展示类内容模块（个人设备 / 项目经历 / 技术栈 / 开源作品）
	s.registerEquipmentTools(server)
	s.registerProjectTools(server)
	s.registerTechStackTools(server)
	s.registerOpenSourceTools(server)

	// 内容创作模块（歌单 / 视频 / 旅行 / 摄影 / 菜谱 / 书 / 游戏 / 健身）
	s.registerPlaylistTools(server)
	s.registerVideoTools(server)
	s.registerTravelTools(server)
	s.registerPortfolioTools(server)
	s.registerRecipeTools(server)
	s.registerBookTools(server)
	s.registerGameTools(server)
	s.registerFitnessTools(server)
}

// wrapTool 包装工具 handler：统一记录调用日志（密钥、工具名、耗时、结果）。
func wrapTool[In, Out any](s *Service, name string, h sdk.ToolHandlerFor[In, Out]) sdk.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *sdk.CallToolRequest, in In) (result *sdk.CallToolResult, out Out, err error) {
		start := time.Now()
		defer func() {
			fields := []zap.Field{
				zap.String("module", "mcp"),
				zap.String("tool", name),
				zap.String("key_id", KeyIDFromContext(ctx)),
				zap.Int64("cost_ms", time.Since(start).Milliseconds()),
			}
			if err != nil {
				fields = append(fields, zap.Error(err))
				s.logger.Warn("MCP 工具调用失败", fields...)
			} else if recoverErr := recover(); recoverErr != nil {
				s.logger.Error("MCP 工具调用 panic", append(fields, zap.Any("panic", recoverErr))...)
				panic(recoverErr)
			} else {
				s.logger.Info("MCP 工具调用成功", fields...)
			}
		}()
		return h(ctx, req, in)
	}
}

// 文章状态与内部状态码（1 草稿 / 2 已发布 / 3 已下架）的映射。
const (
	statusDraft     = "draft"
	statusPublished = "published"
	statusOffline   = "offline"
)

// statusToString 将内部状态码转为工具入参使用的字符串。
func statusToString(s int16) string {
	switch s {
	case 2:
		return statusPublished
	case 3:
		return statusOffline
	default:
		return statusDraft
	}
}

// statusFromString 将工具入参的状态字符串转为内部状态码，空串返回 0（由调用方决定默认值）。
// 兼容 publish（发布）与 unpublish（下架）两个口语化别名。
func statusFromString(s string) (int16, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return 0, nil
	case statusDraft:
		return 1, nil
	case statusPublished, "publish":
		return 2, nil
	case statusOffline, "unpublish":
		return 3, nil
	default:
		return 0, fmt.Errorf("非法的状态值 %q，可选：draft / published / offline", s)
	}
}

// truncateSummary 从正文生成默认摘要（去除 Markdown 标记后截取前 200 字符）。
func truncateSummary(content string) string {
	plain := strings.Map(func(r rune) rune {
		switch r {
		case '#', '*', '`', '>', '-', '!', '[', ']', '(', ')', '\n', '\r':
			return ' '
		}
		return r
	}, content)
	plain = strings.Join(strings.Fields(plain), " ")
	runes := []rune(plain)
	if len(runes) > 200 {
		plain = string(runes[:200]) + "…"
	}
	return plain
}
