package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"novablog/enum"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/githubartifact"
	"novablog/pkg/novablogapi"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ThemeLogic 主题安装/激活/卸载/列表编排。
type ThemeLogic struct {
	themeModel   *model.ThemeModel
	bloggerModel *model.BloggerModel
	artifact     *ThemeArtifactLogic
	github       *githubartifact.Client
}

// NewThemeLogic 创建 ThemeLogic 实例。
func NewThemeLogic() *ThemeLogic {
	settings := getThemeSettings()
	return &ThemeLogic{
		themeModel:   model.NewTheme(),
		bloggerModel: model.NewBlogger(),
		artifact:     NewThemeArtifactLogic(),
		github:       &githubartifact.Client{Token: settings.GithubToken, ProxyURL: settings.ProxyURL},
	}
}

// List 已安装主题列表（含激活标记）。
func (l *ThemeLogic) List(ctx context.Context) ([]res.ThemeItemRes, error) {
	themes, err := l.themeModel.List(ctx)
	if err != nil {
		return nil, err
	}

	activeID := ""
	if blogger, err := l.bloggerModel.GetFirst(ctx); err == nil && blogger.ActiveThemeID != nil {
		activeID = *blogger.ActiveThemeID
	}

	items := make([]res.ThemeItemRes, 0, len(themes))
	for _, t := range themes {
		items = append(items, mapThemeRes(t, activeID))
	}
	return items, nil
}

// Install 从官方市场安装指定主题（按市场主题 ID，可指定版本）。
// marketBaseURL 优先使用传入值，为空时回退 config.yaml 的 themes.market_base_url。
func (l *ThemeLogic) Install(ctx context.Context, marketID int64, version string, force bool, marketBaseURL string) (*model.Theme, error) {
	baseURL, err := resolveMarketBaseURL(marketBaseURL)
	if err != nil {
		return nil, err
	}
	item, err := novablogapi.GetTheme(ctx, baseURL, strconv.FormatInt(marketID, 10), "")
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return l.installFromMarketItem(ctx, item.ThemeItem, item.DownloadURL, version, force)
}

// InstallDefault 首装链路：从官方市场拉取默认主题安装。
// 市场未配置、不可达或无默认主题时抛明确错误（显示官方地址不可达），无兜底。
func (l *ThemeLogic) InstallDefault(ctx context.Context) (*model.Theme, error) {
	settings := getThemeSettings()
	if settings.MarketBaseURL == "" {
		return nil, enum.ErrThemeMarketNotConfigured
	}
	base, err := novablogapi.NormalizeBaseURL(settings.MarketBaseURL)
	if err != nil {
		return nil, enum.ErrThemeMarketNotConfigured
	}

	dt, err := novablogapi.GetDefaultTheme(ctx, base)
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	theme, installErr := l.installFromMarketItem(ctx, dt.ThemeItem, dt.DownloadURL, "", false)
	if installErr != nil {
		return nil, installErr
	}
	return theme, nil
}

// installFromMarketItem 从市场主题元数据执行完整安装链路：
// 解析制品地址 → 下载+sha256 → 安全解压 → 清单校验 → 注入配置 → 原子落盘 → 落库。
//
// download_url 支持两种形式：
//   - 直接 tar.gz 地址（以 .tar.gz 结尾）：直接下载，跳过 GitHub Release 解析
//   - GitHub 仓库目录地址（形如 https://github.com/{owner}/{repo}/tree/{ref}/{dir}）：
//     走旧逻辑，查 GitHub Release 找 {dir}-{version}.tar.gz 制品
func (l *ThemeLogic) installFromMarketItem(ctx context.Context, item novablogapi.ThemeItem, downloadURL, version string, force bool) (*model.Theme, error) {
	if version == "" {
		version = item.Version
	}

	// 幂等：同主题同版本已安装
	existing, existingErr := l.themeModel.GetByThemeAndVersion(ctx, item.Slug, version)
	if existingErr == nil && !force {
		return existing, nil
	}
	if existingErr == nil && force {
		// 覆盖安装：激活中的实例不允许覆盖（先切换其他主题）
		if blogger, berr := l.bloggerModel.GetFirst(ctx); berr == nil && blogger.ActiveThemeID != nil && *blogger.ActiveThemeID == existing.ID {
			return nil, enum.ErrThemeActiveCannotDelete
		}
	}

	var (
		assetURL         string  // 实际下载地址
		dir              string  // 主题目录名（用于清单校验）
		owner, repo, tag string  // 旧逻辑的 GitHub 元数据（用于 sourceRef）
		useGitHub        = false // 是否走 GitHub Release 解析
	)

	if strings.HasSuffix(downloadURL, ".tar.gz") {
		// ① 直接 tar.gz 地址：从文件名解析 dir 和 version
		assetURL = downloadURL
		filename := filepath.Base(downloadURL) // 如 "tech-geek-0.1.0.tar.gz"
		parsedDir, parsedVersion := parseDirVersion(filename)
		if parsedDir == "" {
			return nil, enum.ErrThemeArtifactNotFound
		}
		dir = parsedDir
		if parsedVersion != "" {
			version = parsedVersion
		}
	} else {
		// ② GitHub 仓库目录地址 → 查 Release 制品（旧逻辑）
		useGitHub = true
		var err error
		owner, repo, dir, err = githubartifact.ParseRepoDir(downloadURL)
		if err != nil {
			return nil, enum.ErrThemeArtifactNotFound
		}
		asset, err := l.github.FindReleaseAsset(ctx, owner, repo, dir, version)
		if err != nil {
			if errors.Is(err, githubartifact.ErrArtifactNotFound) || errors.Is(err, githubartifact.ErrInvalidRepoURL) {
				return nil, enum.ErrThemeArtifactNotFound
			}
			return nil, enum.NewBizError(enum.ErrMarketUpstream.Code, err.Error(), enum.ErrMarketUpstream.HttpCode)
		}
		assetURL = asset.DownloadURL
		tag = asset.Tag
	}

	// 流式下载（边下边算 sha256）
	archivePath, checksum, err := l.artifact.DownloadArtifact(ctx, assetURL)
	if err != nil {
		return nil, err
	}
	defer os.Remove(archivePath)

	// 安全解压 + 清单校验
	tmpRoot, err := l.artifact.TempRoot()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpRoot)

	if err := l.artifact.SafeExtract(archivePath, tmpRoot); err != nil {
		return nil, err
	}
	manifest, err := l.artifact.LoadManifest(tmpRoot)
	if err != nil {
		return nil, err
	}
	if manifest.ID != dir {
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "清单 id 与制品目录不一致", enum.ErrThemeManifestInvalid.HttpCode)
	}
	if version != "" && manifest.Version != version {
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "制品版本与请求版本不一致", enum.ErrThemeManifestInvalid.HttpCode)
	}

	// 注入运行时配置（跨域场景）
	if err := l.artifact.InjectThemeConfig(tmpRoot, getThemeSettings().PublicAPIBase); err != nil {
		return nil, err
	}

	// 覆盖安装：先清理旧目录与旧记录
	if existingErr == nil && force {
		_ = os.RemoveAll(filepath.Join(getThemeSettings().DataDir, existing.ArtifactPath))
		_ = l.themeModel.Delete(ctx, existing.ID)
	}

	rel, err := l.artifact.PlaceArtifact(tmpRoot, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}

	// sourceRef 构造
	var sourceRef string
	if useGitHub {
		sourceRef = fmt.Sprintf("%s/%s@%s", owner, repo, tag)
	} else {
		sourceRef = downloadURL
	}
	return l.createThemeRecord(ctx, manifest, "official", sourceRef, item.ID, item.Slug, rel, checksum)
}

// UpdateTheme 从官方市场更新已安装主题到最新版本。
// marketBaseURL 优先使用传入值，为空时回退 config.yaml 的 themes.market_base_url。
func (l *ThemeLogic) UpdateTheme(ctx context.Context, id string, marketBaseURL string) (*model.Theme, error) {
	// 1. 获取已安装主题
	theme, err := l.themeModel.GetByID(ctx, id)
	if err != nil {
		return nil, enum.ErrThemeNotFound
	}

	// 2. 仅官方来源的主题可更新（需有 MarketID）
	if theme.MarketID <= 0 {
		return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "非官方来源的主题无法检查更新", enum.ErrInvalidParam.HttpCode)
	}

	// 3. 解析市场地址
	baseURL, err := resolveMarketBaseURL(marketBaseURL)
	if err != nil {
		return nil, err
	}

	// 4. 从官方市场拉取最新版本详情
	marketID := strconv.FormatInt(theme.MarketID, 10)
	detail, err := novablogapi.GetTheme(ctx, baseURL, marketID, "")
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	// 5. 比对版本号
	if detail.Version == "" {
		return nil, enum.ErrMarketUpstream
	}
	if semverCompare(detail.Version, theme.Version) <= 0 {
		return nil, enum.ErrAlreadyLatestTheme
	}

	// 6. 安装新版本
	upgraded, err := l.installFromMarketItem(ctx, detail.ThemeItem, detail.DownloadURL, detail.Version, false)
	if err != nil {
		return nil, err
	}

	// 7. 旧版本正在使用中 → 自动激活新版本
	blogger, berr := l.bloggerModel.GetFirst(ctx)
	if berr == nil && blogger.ActiveThemeID != nil && *blogger.ActiveThemeID == theme.ID {
		if actErr := l.bloggerModel.UpdateActiveTheme(ctx, blogger.ID, upgraded.ID); actErr != nil {
			themeLog().Warn("更新后自动激活新版本失败", zap.Error(actErr))
		} else {
			ThemeHost.Invalidate()
			themeLog().Info("主题更新后已自动切换激活",
				zap.String("theme_id", upgraded.ThemeID),
				zap.String("new_version", upgraded.Version),
			)
		}
	}

	// 8. 清理旧版本制品与记录
	oldArtifactPath := filepath.Join(getThemeSettings().DataDir, theme.ArtifactPath)
	if rmErr := os.RemoveAll(oldArtifactPath); rmErr != nil {
		themeLog().Warn("清理旧主题制品目录失败", zap.String("path", theme.ArtifactPath), zap.Error(rmErr))
	}
	if delErr := l.themeModel.Delete(ctx, theme.ID); delErr != nil {
		themeLog().Warn("删除旧主题记录失败", zap.Error(delErr))
	}

	themeLog().Info("主题更新完成",
		zap.String("theme_id", upgraded.ThemeID),
		zap.String("old_version", theme.Version),
		zap.String("new_version", upgraded.Version),
	)
	return upgraded, nil
}

// semverCompare 简单语义化版本比较，支持 x.y.z 格式及带 pre-release 后缀。
// 返回 -1（a < b）, 0（a == b）, 1（a > b）。
func semverCompare(a, b string) int {
	parse := func(v string) (major, minor, patch int, pre string) {
		v = strings.TrimLeft(v, "v")
		if idx := strings.IndexAny(v, "-+"); idx != -1 {
			pre = v[idx:]
			v = v[:idx]
		}
		parts := strings.Split(v, ".")
		if len(parts) != 3 {
			return 0, 0, 0, ""
		}
		major, _ = strconv.Atoi(parts[0])
		minor, _ = strconv.Atoi(parts[1])
		patch, _ = strconv.Atoi(parts[2])
		return
	}

	ma, na, pa, preA := parse(a)
	mb, nb, pb, preB := parse(b)

	switch {
	case ma != mb:
		return sign(ma - mb)
	case na != nb:
		return sign(na - nb)
	case pa != pb:
		return sign(pa - pb)
	}

	// pre-release 比较：无前缀 > 有前缀
	switch {
	case preA == preB:
		return 0
	case preA == "":
		return 1
	case preB == "":
		return -1
	default:
		if preA < preB {
			return -1
		}
		return 1
	}
}

// sign 返回整数的正负号（-1/0/1）。
func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// resolveMarketBaseURL 解析官方市场地址：优先使用传入值，为空时回退 config。
// 返回归一化后的 baseURL（含 /api/v1 前缀），或 ErrThemeMarketNotConfigured。
func resolveMarketBaseURL(marketBaseURL string) (string, error) {
	if marketBaseURL != "" {
		base, err := novablogapi.NormalizeBaseURL(marketBaseURL)
		if err == nil {
			return base, nil
		}
	}
	settings := getThemeSettings()
	if settings.MarketBaseURL == "" {
		return "", enum.ErrThemeMarketNotConfigured
	}
	base, err := novablogapi.NormalizeBaseURL(settings.MarketBaseURL)
	if err != nil {
		return "", enum.ErrThemeMarketNotConfigured
	}
	return base, nil
}

// parseDirVersion 从 tar.gz 文件名中解析主题目录名和版本。
// 文件名格式: {dir}-{version}.tar.gz，如 "tech-geek-0.1.0.tar.gz" → "tech-geek", "0.1.0"
// 解析失败时返回空字符串。
func parseDirVersion(filename string) (dir, version string) {
	m := assetNamePattern.FindStringSubmatch(filename)
	if m != nil {
		return m[1], m[2]
	}
	// 兼容无版本号的情况：直接去掉 .tar.gz 后缀
	trimmed := strings.TrimSuffix(filename, ".tar.gz")
	if trimmed != "" && trimmed != filename {
		return trimmed, ""
	}
	return "", ""
}

// assetNamePattern 匹配 {dir}-{version}.tar.gz 并提取目录和版本。
var assetNamePattern = regexp.MustCompile(`^(.+)-(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)\.tar\.gz$`)

// createThemeRecord 写入主题安装记录。
func (l *ThemeLogic) createThemeRecord(ctx context.Context, manifest *ThemeManifest, source, sourceRef string, marketID int64, marketSlug, artifactPath, checksum string) (*model.Theme, error) {
	screenshots, _ := json.Marshal(manifest.Screenshots)
	fallbacks, _ := json.Marshal(manifest.Routes.Fallback)
	if manifest.Routes.Fallback == nil {
		fallbacks = []byte("{}")
	}

	record := &model.Theme{
		ID:           uuid.New().String(),
		ThemeID:      manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Engine:       manifest.Engine,
		APICompat:    manifest.APICompat,
		Author:       manifest.Author,
		Description:  manifest.Description,
		Screenshots:  string(screenshots),
		Fallbacks:    string(fallbacks),
		Source:       source,
		SourceRef:    sourceRef,
		MarketID:     marketID,
		MarketSlug:   marketSlug,
		ArtifactPath: artifactPath,
		Checksum:     checksum,
	}
	if err := l.themeModel.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("保存主题安装记录失败: %w", err)
	}
	themeLog().Info("主题安装成功",
		zap.String("theme_id", record.ThemeID),
		zap.String("version", record.Version),
		zap.String("source", record.Source),
		zap.String("checksum", record.Checksum),
	)
	return record, nil
}

// Activate 激活主题实例：更新博主激活指针并失效托管缓存（秒级生效）。
func (l *ThemeLogic) Activate(ctx context.Context, id string) (*model.Theme, error) {
	theme, err := l.themeModel.GetByID(ctx, id)
	if err != nil {
		return nil, enum.ErrThemeNotFound
	}

	blogger, err := l.bloggerModel.GetFirst(ctx)
	if err != nil {
		return nil, enum.ErrInternalServer
	}
	if err := l.bloggerModel.UpdateActiveTheme(ctx, blogger.ID, theme.ID); err != nil {
		return nil, fmt.Errorf("更新激活指针失败: %w", err)
	}

	ThemeHost.Invalidate()
	themeLog().Info("主题已激活", zap.String("theme_id", theme.ThemeID), zap.String("version", theme.Version))
	return theme, nil
}

// Uninstall 卸载主题实例（激活中的实例拒绝卸载）。
func (l *ThemeLogic) Uninstall(ctx context.Context, id string) error {
	theme, err := l.themeModel.GetByID(ctx, id)
	if err != nil {
		return enum.ErrThemeNotFound
	}

	if blogger, err := l.bloggerModel.GetFirst(ctx); err == nil && blogger.ActiveThemeID != nil && *blogger.ActiveThemeID == id {
		return enum.ErrThemeActiveCannotDelete
	}

	if err := os.RemoveAll(filepath.Join(getThemeSettings().DataDir, theme.ArtifactPath)); err != nil {
		themeLog().Warn("清理主题目录失败", zap.String("path", theme.ArtifactPath), zap.Error(err))
	}
	if err := l.themeModel.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除主题记录失败: %w", err)
	}

	ThemeHost.Invalidate()
	themeLog().Info("主题已卸载", zap.String("theme_id", theme.ThemeID), zap.String("version", theme.Version))
	return nil
}

// GetActiveTheme 当前激活主题实例（未激活或实例已丢失时返回 nil）。
func (l *ThemeLogic) GetActiveTheme(ctx context.Context) (*model.Theme, error) {
	blogger, err := l.bloggerModel.GetFirst(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if blogger.ActiveThemeID == nil || *blogger.ActiveThemeID == "" {
		return nil, nil
	}
	theme, err := l.themeModel.GetByID(ctx, *blogger.ActiveThemeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return theme, nil
}

// mapThemeRes 主题安装记录 → 响应 DTO。
func mapThemeRes(t model.Theme, activeID string) res.ThemeItemRes {
	var screenshots []string
	_ = json.Unmarshal([]byte(t.Screenshots), &screenshots)
	if screenshots == nil {
		screenshots = []string{}
	}
	fallbacks := map[string]string{}
	_ = json.Unmarshal([]byte(t.Fallbacks), &fallbacks)
	if fallbacks == nil {
		fallbacks = map[string]string{}
	}
	return res.ThemeItemRes{
		ID:           t.ID,
		ThemeID:      t.ThemeID,
		Name:         t.Name,
		Version:      t.Version,
		Engine:       t.Engine,
		APICompat:    t.APICompat,
		Author:       t.Author,
		Description:  t.Description,
		Source:       t.Source,
		Screenshots:  screenshots,
		Fallbacks:    fallbacks,
		MarketID:     t.MarketID,
		MarketSlug:   t.MarketSlug,
		ArtifactPath: t.ArtifactPath,
		Checksum:     t.Checksum,
		Active:       t.ID == activeID,
		CreatedAt:    t.CreatedAt,
	}
}
