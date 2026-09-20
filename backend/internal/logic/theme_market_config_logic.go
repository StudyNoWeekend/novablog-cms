package logic

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/novablogapi"
)

// themeMarketConfigCacheTTL 主题模块运行配置缓存过期时间。
const themeMarketConfigCacheTTL = 10 * time.Minute

// ThemeMarketConfigLogic 主题模块运行配置业务逻辑结构体。
// 官方市场地址与博客公开 API 地址的持久化统一入口：首装向导 / 市场登录 / 后台修改
// 均落到 DB 单行表并同步内存单例热更新；config.yaml 的对应字段仅作出厂默认值。
type ThemeMarketConfigLogic struct {
	configModel *model.ThemeMarketConfigModel
	configCache *cache.ThemeMarketConfigCache
}

// NewThemeMarketConfigLogic 创建 ThemeMarketConfigLogic 实例。
func NewThemeMarketConfigLogic() *ThemeMarketConfigLogic {
	return &ThemeMarketConfigLogic{
		configModel: model.NewThemeMarketConfig(),
		configCache: cache.NewThemeMarketConfigCache(),
	}
}

// GetConfig 获取主题模块运行配置（优先读缓存，未命中查库并回填缓存）。
// 字段为空表示未自定义，实际生效值由调用方回退 config.yaml 出厂值。
func (l *ThemeMarketConfigLogic) GetConfig(ctx context.Context) (*res.ThemeMarketConfigRes, error) {
	cached, err := l.configCache.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取主题运行配置缓存失败: %w", err)
	}
	if cached != nil {
		return &res.ThemeMarketConfigRes{
			MarketBaseURL: cached.MarketBaseURL,
			PublicAPIBase: cached.PublicAPIBase,
			UpdatedAt:     cached.UpdatedAt,
		}, nil
	}

	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询主题运行配置失败: %w", err)
	}

	_ = l.configCache.SetConfig(ctx, &cache.ThemeMarketConfigCacheData{
		MarketBaseURL: config.MarketBaseURL,
		PublicAPIBase: config.PublicAPIBase,
	}, themeMarketConfigCacheTTL)

	return &res.ThemeMarketConfigRes{
		MarketBaseURL: config.MarketBaseURL,
		PublicAPIBase: config.PublicAPIBase,
		UpdatedAt:     config.UpdatedAt,
	}, nil
}

// GetPublicConfig 公共下发配置：返回当前生效的官方地址（DB 自定义值优先，回退出厂值）。
func (l *ThemeMarketConfigLogic) GetPublicConfig(ctx context.Context) (*res.PublicConfigRes, error) {
	config, err := l.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if config.MarketBaseURL != "" {
		return &res.PublicConfigRes{MarketBaseURL: config.MarketBaseURL}, nil
	}
	return &res.PublicConfigRes{MarketBaseURL: getThemeSettings().MarketBaseURL}, nil
}

// SyncMarketBaseURL 持久化官方地址并热更新内存单例（首装向导 / 市场登录 / 后台修改共用入口）。
// raw 为空时跳过；地址归一化（自动补 /api/v1）后落库。
func (l *ThemeMarketConfigLogic) SyncMarketBaseURL(ctx context.Context, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	base, err := novablogapi.NormalizeBaseURL(raw)
	if err != nil {
		return enum.NewBizError(enum.ErrMarketBaseURLInvalid.Code, err.Error(), enum.ErrMarketBaseURLInvalid.HttpCode)
	}

	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询主题运行配置失败: %w", err)
	}
	config.MarketBaseURL = base
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新主题运行配置失败: %w", err)
	}
	_ = l.configCache.SetConfig(ctx, &cache.ThemeMarketConfigCacheData{
		MarketBaseURL: base,
		PublicAPIBase: config.PublicAPIBase,
		UpdatedAt:     time.Now(),
	}, themeMarketConfigCacheTTL)

	// 同步内存单例，市场代理与首装拉取立即生效
	getThemeSettings().MarketBaseURL = base
	return nil
}

// syncPublicAPIBase 持久化博客公开 API 地址并热更新内存单例，随后对已安装主题重注入。
// raw 为空表示清除（回退同域相对路径取数）；非空需为合法 http(s) 地址。
func (l *ThemeMarketConfigLogic) syncPublicAPIBase(ctx context.Context, raw string) error {
	base := strings.TrimSpace(raw)
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		return enum.NewBizError(enum.ErrInvalidParam.Code, "博客 API 地址需以 http:// 或 https:// 开头", enum.ErrInvalidParam.HttpCode)
	}

	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("查询主题运行配置失败: %w", err)
	}
	config.PublicAPIBase = base
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新主题运行配置失败: %w", err)
	}
	_ = l.configCache.SetConfig(ctx, &cache.ThemeMarketConfigCacheData{
		MarketBaseURL: config.MarketBaseURL,
		PublicAPIBase: base,
		UpdatedAt:     time.Now(),
	}, themeMarketConfigCacheTTL)

	// 同步内存单例：后续主题安装/更新按新地址注入；清空时显式覆盖（区别于启动注入的仅非空语义）
	getThemeSettings().PublicAPIBase = base

	l.reinjectInstalledThemes(ctx, base)
	return nil
}

// reinjectInstalledThemes 对全部已安装主题重写 dist/theme-config.js（含预览中的主题），失败仅记录不影响保存。
func (l *ThemeMarketConfigLogic) reinjectInstalledThemes(ctx context.Context, apiBase string) {
	themes, err := model.NewTheme().List(ctx)
	if err != nil {
		themeLog().Warn("重注入主题配置失败：查询已安装主题出错", zap.Error(err))
		return
	}
	artifact := NewThemeArtifactLogic()
	for i := range themes {
		root := filepath.Join(getThemeSettings().DataDir, themes[i].ArtifactPath)
		if err := artifact.ApplyThemeConfig(root, apiBase); err != nil {
			themeLog().Warn("重注入主题配置失败",
				zap.String("theme", themes[i].ThemeID),
				zap.String("artifact", themes[i].ArtifactPath),
				zap.Error(err))
		}
	}
}

// UpdateConfig 后台更新主题模块运行配置（两个字段均可选，仅更新提供的字段）。
// market_base_url 必须非空且合法；public_api_base 允许空（清除，回退同域取数）。
func (l *ThemeMarketConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateThemeMarketConfigReq) error {
	if r.MarketBaseURL != nil && strings.TrimSpace(*r.MarketBaseURL) == "" {
		return enum.NewBizError(enum.ErrInvalidParam.Code, "官方地址不能为空", enum.ErrInvalidParam.HttpCode)
	}
	if r.MarketBaseURL != nil {
		if err := l.SyncMarketBaseURL(ctx, *r.MarketBaseURL); err != nil {
			return err
		}
	}
	if r.PublicAPIBase != nil {
		if err := l.syncPublicAPIBase(ctx, *r.PublicAPIBase); err != nil {
			return err
		}
	}
	return nil
}
