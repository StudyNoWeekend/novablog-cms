package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/pkg/novablogapi"
)

// themeMarketConfigCacheTTL 官方市场配置缓存过期时间。
const themeMarketConfigCacheTTL = 10 * time.Minute

// ThemeMarketConfigLogic 官方主题市场配置业务逻辑结构体。
// 官方地址的持久化统一入口：首装向导 / 市场登录 / 后台修改均落到 DB 单行表，
// 并同步内存单例热更新；config.yaml 的 themes.market_base_url 仅作出厂默认值。
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

// GetConfig 获取官方市场配置（优先读缓存，未命中查库并回填缓存）。
// MarketBaseURL 为空表示未自定义，实际生效值由调用方回退 config.yaml 出厂值。
func (l *ThemeMarketConfigLogic) GetConfig(ctx context.Context) (*res.ThemeMarketConfigRes, error) {
	cached, err := l.configCache.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取官方市场配置缓存失败: %w", err)
	}
	if cached != nil {
		return &res.ThemeMarketConfigRes{MarketBaseURL: cached.MarketBaseURL, UpdatedAt: cached.UpdatedAt}, nil
	}

	config, err := l.configModel.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询官方市场配置失败: %w", err)
	}

	_ = l.configCache.SetConfig(ctx, &cache.ThemeMarketConfigCacheData{MarketBaseURL: config.MarketBaseURL}, themeMarketConfigCacheTTL)

	return &res.ThemeMarketConfigRes{
		MarketBaseURL: config.MarketBaseURL,
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
		return fmt.Errorf("查询官方市场配置失败: %w", err)
	}
	config.MarketBaseURL = base
	if err := l.configModel.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("更新官方市场配置失败: %w", err)
	}
	_ = l.configCache.SetConfig(ctx, &cache.ThemeMarketConfigCacheData{MarketBaseURL: base, UpdatedAt: time.Now()}, themeMarketConfigCacheTTL)

	// 同步内存单例，市场代理与首装拉取立即生效
	getThemeSettings().MarketBaseURL = base
	return nil
}

// UpdateConfig 后台更新官方地址（必填；归一化后持久化并热更新）。
func (l *ThemeMarketConfigLogic) UpdateConfig(ctx context.Context, r *req.UpdateThemeMarketConfigReq) error {
	if r.MarketBaseURL == nil || strings.TrimSpace(*r.MarketBaseURL) == "" {
		return enum.NewBizError(enum.ErrInvalidParam.Code, "官方地址不能为空", enum.ErrInvalidParam.HttpCode)
	}
	return l.SyncMarketBaseURL(ctx, *r.MarketBaseURL)
}
