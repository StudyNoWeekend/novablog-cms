// Package mcp 实现 NovaBlog 的 MCP（Model Context Protocol）服务端，
// 通过 Streamable HTTP 传输对外暴露文章发布、媒体上传等工具，
// 供 Codex / ZCode 等 AI 客户端远程操作博客内容。
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

// Service MCP 服务端结构体：聚合发布内容所需的业务 logic。
type Service struct {
	version       string
	logger        *zap.Logger
	articleLogic  *logic.ArticleLogic
	mediaLogic    *logic.MediaLogic
	categoryLogic *logic.CategoryLogic
	tagLogic      *logic.TagLogic
}

// NewService 创建 MCP 服务端实例。
func NewService(version string, mediaLogic *logic.MediaLogic, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		version:       version,
		logger:        logger,
		articleLogic:  logic.NewArticleLogic(),
		mediaLogic:    mediaLogic,
		categoryLogic: logic.NewCategoryLogic(),
		tagLogic:      logic.NewTagLogic(),
	}
}

// Handler 返回 MCP Streamable HTTP 处理器（无状态模式，每个 POST 请求独立处理）。
func (s *Service) Handler() http.Handler {
	server := sdk.NewServer(&sdk.Implementation{
		Name:    "novablog",
		Version: s.version,
	}, &sdk.ServerOptions{
		Instructions: "NovaBlog 博客内容发布服务。" +
			"可用的能力：创建/更新/发布/下架/删除文章（正文为 Markdown），上传图片到媒体库（用于封面图和正文配图），查询分类与标签。" +
			"建议流程：先上传封面图与正文图片拿到 URL，再创建文章（默认保存为草稿，明确要求发布时才置为 published）。" +
			"复用已有分类和标签（先 list_categories / list_tags 查询），保持站点内容结构一致。",
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
		Description: "上传图片到媒体库并返回可直接引用的 URL。source 支持 http(s) 图片链接（自动转存）或 base64 图片数据。用于封面图和正文配图。",
	}, wrapTool(s, "upload_image", s.uploadImage))

	// 分类与标签
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_categories",
		Description: "查询文章分类列表，创建文章前先查询以便复用。",
	}, wrapTool(s, "list_categories", s.listCategories))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_tags",
		Description: "查询标签列表，创建文章前先查询以便复用。",
	}, wrapTool(s, "list_tags", s.listTags))
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
