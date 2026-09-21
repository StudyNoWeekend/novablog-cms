package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/githubartifact"

	"github.com/google/uuid"
)

// OpenSourceSettings 开源作品模块配置，由 bootstrap 阶段注入。
// 与主题模块共用 GitHub 访问配置（config.yaml 的 themes.github_token / themes.proxy_url）。
type OpenSourceSettings struct {
	GithubToken string // 可选 GitHub PAT（提升 API 限流额度）
	ProxyURL    string // 可选 HTTP 代理（空=直连；代理不可用自动降级直连）
}

// openSourceSettings 开源作品模块运行时配置。
var openSourceSettings = &OpenSourceSettings{}

// SetOpenSourceSettings 注入开源作品模块配置。
func SetOpenSourceSettings(s *OpenSourceSettings) {
	if s != nil {
		openSourceSettings = s
	}
}

// OpenSourceLogic 开源作品业务逻辑结构体。
type OpenSourceLogic struct {
	workModel *model.OpenSourceWorkModel
}

// NewOpenSourceLogic 创建 OpenSourceLogic 实例。
func NewOpenSourceLogic() *OpenSourceLogic {
	return &OpenSourceLogic{
		workModel: model.NewOpenSourceWork(),
	}
}

// githubClient 按当前配置构建 GitHub 客户端。
func (l *OpenSourceLogic) githubClient() *githubartifact.Client {
	return &githubartifact.Client{
		Token:    openSourceSettings.GithubToken,
		ProxyURL: openSourceSettings.ProxyURL,
	}
}

// Create 创建开源作品。默认自动拉取仓库元数据与 README：
// 元数据回填空缺字段（简介/语言/标签/主页），README 原文入库供公开接口展示。
func (l *OpenSourceLogic) Create(ctx context.Context, r *req.CreateOpenSourceReq) (*res.OpenSourceRes, error) {
	owner, repoName, err := githubartifact.ParseRepoURL(r.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("仓库链接不合法，仅支持 GitHub 仓库地址，如 https://github.com/owner/repo")
	}

	status := 0 // 默认草稿
	if r.Status != nil {
		status = *r.Status
	}
	sortOrder := 0
	if r.SortOrder != nil {
		sortOrder = *r.SortOrder
	}

	w := &model.OpenSourceWork{
		ID:        uuid.New().String(),
		Name:      strings.TrimSpace(r.Name),
		RepoURL:   strings.TrimSpace(r.RepoURL),
		Summary:   r.Summary,
		Language:  r.Language,
		Topics:    r.Topics,
		Homepage:  r.Homepage,
		Status:    status,
		SortOrder: sortOrder,
	}

	if r.FetchRemote == nil || *r.FetchRemote {
		if err := l.syncRemote(ctx, w, owner, repoName); err != nil {
			return nil, err
		}
	}

	if err := l.workModel.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("创建开源作品失败: %w", err)
	}

	return l.toOpenSourceRes(w), nil
}

// GetList 分页查询开源作品列表（不含 README）。
func (l *OpenSourceLogic) GetList(ctx context.Context, r *req.OpenSourceListReq) (*res.OpenSourceListRes, error) {
	list, total, err := l.workModel.GetList(ctx, r.Keyword, r.Status, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, fmt.Errorf("查询开源作品列表失败: %w", err)
	}

	items := make([]res.OpenSourceCardRes, 0, len(list))
	for i := range list {
		items = append(items, l.toOpenSourceCard(&list[i]))
	}

	return &res.OpenSourceListRes{
		List:       items,
		Total:      total,
		Page:       r.GetPage(),
		PageSize:   r.GetPageSize(),
		TotalPages: calcTotalPages(total, r.GetPageSize()),
	}, nil
}

// GetByID 查询开源作品详情（含 README）。
func (l *OpenSourceLogic) GetByID(ctx context.Context, id string) (*res.OpenSourceRes, error) {
	w, err := l.workModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("开源作品不存在")
	}
	return l.toOpenSourceRes(w), nil
}

// Update 更新开源作品（部分更新语义）。仓库链接变更时自动重新拉取元数据与 README。
func (l *OpenSourceLogic) Update(ctx context.Context, id string, r *req.UpdateOpenSourceReq) (*res.OpenSourceRes, error) {
	w, err := l.workModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("开源作品不存在")
	}

	repoChanged := false
	if r.Name != nil {
		w.Name = strings.TrimSpace(*r.Name)
	}
	if r.RepoURL != nil {
		newURL := strings.TrimSpace(*r.RepoURL)
		repoChanged = newURL != w.RepoURL
		w.RepoURL = newURL
	}
	if r.Summary != nil {
		w.Summary = *r.Summary
	}
	if r.Language != nil {
		w.Language = *r.Language
	}
	if r.Topics != nil {
		w.Topics = *r.Topics
	}
	if r.Homepage != nil {
		w.Homepage = *r.Homepage
	}
	if r.Status != nil {
		w.Status = *r.Status
	}
	if r.SortOrder != nil {
		w.SortOrder = *r.SortOrder
	}

	owner, repoName, parseErr := githubartifact.ParseRepoURL(w.RepoURL)
	if parseErr != nil {
		return nil, fmt.Errorf("仓库链接不合法，仅支持 GitHub 仓库地址，如 https://github.com/owner/repo")
	}

	if repoChanged {
		if syncErr := l.syncRemote(ctx, w, owner, repoName); syncErr != nil {
			return nil, syncErr
		}
	}

	if err := l.workModel.Update(ctx, w); err != nil {
		return nil, fmt.Errorf("更新开源作品失败: %w", err)
	}

	return l.toOpenSourceRes(w), nil
}

// Refresh 重新拉取远端仓库元数据与 README：
// README 与 Star 快照始终更新，简介/语言/标签/主页仅在为空时补全（不覆盖手动编辑）。
func (l *OpenSourceLogic) Refresh(ctx context.Context, id string) (*res.OpenSourceRes, error) {
	w, err := l.workModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("开源作品不存在")
	}

	owner, repoName, err := githubartifact.ParseRepoURL(w.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("仓库链接不合法，仅支持 GitHub 仓库地址，如 https://github.com/owner/repo")
	}

	if err := l.syncRemote(ctx, w, owner, repoName); err != nil {
		return nil, err
	}

	if err := l.workModel.Update(ctx, w); err != nil {
		return nil, fmt.Errorf("更新开源作品失败: %w", err)
	}

	return l.toOpenSourceRes(w), nil
}

// Delete 删除开源作品。
func (l *OpenSourceLogic) Delete(ctx context.Context, id string) error {
	if _, err := l.workModel.GetByID(ctx, id); err != nil {
		return fmt.Errorf("开源作品不存在")
	}

	if err := l.workModel.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("删除开源作品失败: %w", err)
	}

	return nil
}

// GetPublicList 获取已发布开源作品列表（强制 status=1）。
func (l *OpenSourceLogic) GetPublicList(ctx context.Context, r *req.OpenSourceListReq) (*res.OpenSourceListRes, error) {
	published := 1
	r.Status = &published
	return l.GetList(ctx, r)
}

// GetPublicDetail 获取已发布开源作品详情（含 README，未发布视为不存在）。
func (l *OpenSourceLogic) GetPublicDetail(ctx context.Context, id string) (*res.OpenSourceRes, error) {
	w, err := l.workModel.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("开源作品不存在")
	}
	if w.Status != 1 {
		return nil, fmt.Errorf("开源作品不存在")
	}
	return l.toOpenSourceRes(w), nil
}

// syncRemote 拉取远端仓库元数据与 README 写入作品记录：
// README/Star 始终覆盖，其余字段仅在为空时回填。
func (l *OpenSourceLogic) syncRemote(ctx context.Context, w *model.OpenSourceWork, owner, repoName string) error {
	client := l.githubClient()

	readme, err := client.FetchReadme(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("拉取 README 失败: %w", err)
	}
	w.Readme = readme
	now := time.Now()
	w.ReadmeUpdatedAt = &now

	meta, err := client.FetchRepoMeta(ctx, owner, repoName)
	if err != nil {
		// README 已拿到：元数据失败不阻断保存，仅放弃回填
		meta = nil
	}
	if meta != nil {
		w.Stars = meta.Stars
		if w.Summary == "" && meta.Description != "" {
			w.Summary = truncateRunes(meta.Description, 500)
		}
		if w.Language == "" && meta.Language != "" {
			w.Language = meta.Language
		}
		if w.Topics == "" && len(meta.Topics) > 0 {
			w.Topics = strings.Join(meta.Topics, ",")
		}
		if w.Homepage == "" && meta.Homepage != "" {
			w.Homepage = meta.Homepage
		}
	}

	return nil
}

// truncateRunes 按 rune 截断字符串到 max 长度。
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// toOpenSourceRes 转换为开源作品详情响应。
func (l *OpenSourceLogic) toOpenSourceRes(w *model.OpenSourceWork) *res.OpenSourceRes {
	return &res.OpenSourceRes{
		ID:              w.ID,
		Name:            w.Name,
		RepoURL:         w.RepoURL,
		Summary:         w.Summary,
		Readme:          w.Readme,
		Language:        w.Language,
		Topics:          w.Topics,
		Stars:           w.Stars,
		Homepage:        w.Homepage,
		Status:          w.Status,
		SortOrder:       w.SortOrder,
		ReadmeUpdatedAt: w.ReadmeUpdatedAt,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	}
}

// toOpenSourceCard 转换为开源作品列表卡片响应（不含 README）。
func (l *OpenSourceLogic) toOpenSourceCard(w *model.OpenSourceWork) res.OpenSourceCardRes {
	return res.OpenSourceCardRes{
		ID:        w.ID,
		Name:      w.Name,
		RepoURL:   w.RepoURL,
		Summary:   w.Summary,
		Language:  w.Language,
		Topics:    w.Topics,
		Stars:     w.Stars,
		Homepage:  w.Homepage,
		Status:    w.Status,
		SortOrder: w.SortOrder,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
