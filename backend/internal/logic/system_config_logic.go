package logic

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// 系统配置内置默认值（DB 与 yaml 均无值时兜底）。
const (
	defaultMaxArtifactMB    = 100
	defaultCloudMaxFileMB   = 2048
	defaultCloudBatchLimit  = 200
	defaultOfficialSiteBase = "http://api.novablog.cn"
)

// systemConfigLogic 系统配置逻辑单例（进程内一份，含内存缓存）。
var (
	systemConfigOnce     sync.Once
	systemConfigInstance *SystemConfigLogic
)

// SystemConfigLogic 通用系统配置业务逻辑：数据库为运行时唯一主源，
// 进程内缓存启动时全量加载、保存时同步更新（单实例部署，无需分布式失效）。
type SystemConfigLogic struct {
	model *model.SystemConfigModel

	mu    sync.RWMutex
	cache map[string]string // 键值缓存；nil=尚未加载
}

// NewSystemConfig 返回系统配置逻辑单例。
func NewSystemConfig() *SystemConfigLogic {
	systemConfigOnce.Do(func() {
		systemConfigInstance = &SystemConfigLogic{model: model.NewSystemConfig()}
	})
	return systemConfigInstance
}

// loadAll 全量加载缓存（首次访问时调用；DB 失败保持 nil 以便下次重试）。
func (l *SystemConfigLogic) loadAll(ctx context.Context) {
	list, err := l.model.GetAll(ctx)
	if err != nil {
		return
	}
	m := make(map[string]string, len(list))
	for _, item := range list {
		m[item.Key] = item.Value
	}
	l.mu.Lock()
	l.cache = m
	l.mu.Unlock()
}

// get 读取键值（缓存优先；未加载时全量加载一次）。
func (l *SystemConfigLogic) get(ctx context.Context, key string) string {
	l.mu.RLock()
	cache := l.cache
	l.mu.RUnlock()
	if cache == nil {
		l.loadAll(ctx)
		l.mu.RLock()
		cache = l.cache
		l.mu.RUnlock()
	}
	if cache == nil {
		return ""
	}
	return cache[key]
}

// set 写入单键（upsert DB 后同步缓存）。
func (l *SystemConfigLogic) set(ctx context.Context, key, value, description string) error {
	if err := l.model.Upsert(ctx, key, value, description); err != nil {
		return err
	}
	l.mu.Lock()
	if l.cache != nil {
		l.cache[key] = value
	}
	l.mu.Unlock()
	return nil
}

// —— 类型化访问器（消费方统一入口，空值返回内置默认） ——

// GetSiteURL 站点对外地址（媒体 URL 基址权威值；空=沿用 yaml upload.base_url 首装默认）。
func (l *SystemConfigLogic) GetSiteURL(ctx context.Context) string {
	return strings.TrimRight(strings.TrimSpace(l.get(ctx, model.SystemConfigKeySiteURL)), "/")
}

// GetProxyURL 出网 HTTP 代理（主题/开源作品/图片搜索/云盘/geoip 共用；空=直连）。
func (l *SystemConfigLogic) GetProxyURL(ctx context.Context) string {
	return strings.TrimSpace(l.get(ctx, model.SystemConfigKeyProxyURL))
}

// GetGitHubToken GitHub 访问令牌（空=匿名限流额度）。
func (l *SystemConfigLogic) GetGitHubToken(ctx context.Context) string {
	return strings.TrimSpace(l.get(ctx, model.SystemConfigKeyGithubToken))
}

// GetMaxArtifactMB 主题制品下载上限（MB）。
func (l *SystemConfigLogic) GetMaxArtifactMB(ctx context.Context) int {
	return atoiOr(l.get(ctx, model.SystemConfigKeyMaxArtifactMB), defaultMaxArtifactMB)
}

// GetCloudMaxFileSizeMB 云盘导入单文件上限（MB）。
func (l *SystemConfigLogic) GetCloudMaxFileSizeMB(ctx context.Context) int {
	return atoiOr(l.get(ctx, model.SystemConfigKeyCloudMaxFileSizeMB), defaultCloudMaxFileMB)
}

// GetCloudImportBatchLimit 云盘导入单批文件数上限。
func (l *SystemConfigLogic) GetCloudImportBatchLimit(ctx context.Context) int {
	return atoiOr(l.get(ctx, model.SystemConfigKeyCloudImportBatchLimit), defaultCloudBatchLimit)
}

// GetGeoipDownloadURL geoip 离线库下载地址（空=内置默认源）。
func (l *SystemConfigLogic) GetGeoipDownloadURL(ctx context.Context) string {
	return strings.TrimSpace(l.get(ctx, model.SystemConfigKeyGeoipDownloadURL))
}

// DefaultOfficialSiteBase 官方站点地址内置默认（DB 与播种值均空时的最终兜底）。
func DefaultOfficialSiteBase() string {
	return defaultOfficialSiteBase
}

func atoiOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// —— 管理端读写 ——

// systemConfigField 站点信息分组可编辑字段定义。
type systemConfigField struct {
	key         string
	description string
	multiline   bool
}

// systemConfigFields 站点信息分组暴露给管理端的字段（顺序即展示顺序）。
var systemConfigFields = []systemConfigField{
	{model.SystemConfigKeySiteURL, "站点对外地址（媒体文件 URL 基址；修改后请检查跨域白名单是否包含该地址）", false},
	{model.SystemConfigKeyProxyURL, "出网 HTTP 代理（主题下载、图片搜索、云盘、IP 库更新共用；留空=直连）", false},
	{model.SystemConfigKeyGithubToken, "GitHub 访问令牌（可选，提升主题与开源作品的接口限流额度）", false},
	{model.SystemConfigKeyMaxArtifactMB, "主题制品下载上限（MB）", false},
	{model.SystemConfigKeyCloudMaxFileSizeMB, "云盘导入单文件上限（MB）", false},
	{model.SystemConfigKeyCloudImportBatchLimit, "云盘导入单批文件数上限", false},
	{model.SystemConfigKeyGeoipDownloadURL, "IP 离线库下载地址（留空=内置默认源）", false},
}

// GetSiteSettings 读取站点信息分组（字段按 systemConfigFields 顺序返回，未设置返回空串）。
func (l *SystemConfigLogic) GetSiteSettings(ctx context.Context) ([]model.SystemConfig, error) {
	result := make([]model.SystemConfig, 0, len(systemConfigFields))
	for _, f := range systemConfigFields {
		value := l.get(ctx, f.key)
		result = append(result, model.SystemConfig{Key: f.key, Value: value, Description: f.description})
	}
	return result, nil
}

// UpdateSiteSettings 批量更新站点信息分组（nil 或 nil 指针表示不修改该项）。
// siteURL 变化时通过 onSiteURLChange 回调通知调用方热更（如 model.BaseURL）。
func (l *SystemConfigLogic) UpdateSiteSettings(ctx context.Context, values map[string]*string) (siteURLChanged bool, err error) {
	valid := make(map[string]bool, len(systemConfigFields))
	for _, f := range systemConfigFields {
		valid[f.key] = true
	}
	for key, value := range values {
		if !valid[key] || value == nil {
			continue
		}
		newValue := strings.TrimSpace(*value)
		if key == model.SystemConfigKeySiteURL {
			newValue = strings.TrimRight(newValue, "/")
			if newValue != l.GetSiteURL(ctx) {
				siteURLChanged = true
			}
		}
		if err := l.set(ctx, key, newValue, descriptionOf(key)); err != nil {
			return siteURLChanged, fmt.Errorf("更新配置项 %s 失败: %w", key, err)
		}
	}
	return siteURLChanged, nil
}

func descriptionOf(key string) string {
	for _, f := range systemConfigFields {
		if f.key == key {
			return f.description
		}
	}
	return ""
}

// —— 导出 / 导入 ——

// 敏感项脱敏占位（导出默认不携带真实值，导入时跳过该占位）。
const secretPlaceholder = "__REDACTED__"

// BuildConfigExport 汇总全部运行时配置生成交互用导出结构。
// includeSecrets=false 时凭据类字段以占位符输出（导入时自动跳过，不覆盖现有值）。
func (l *SystemConfigLogic) BuildConfigExport(ctx context.Context, includeSecrets bool) (*res.SystemConfigExportRes, error) {
	groups := make(map[string]map[string]string)
	put := func(group string, items map[string]string) {
		filtered := make(map[string]string, len(items))
		for k, v := range items {
			if !includeSecrets && isSecretField(group, k) {
				if v == "" {
					continue
				}
				filtered[k] = secretPlaceholder
				continue
			}
			filtered[k] = v
		}
		groups[group] = filtered
	}

	// 通用系统配置
	list, err := l.model.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	systemItems := make(map[string]string, len(list))
	for _, item := range list {
		systemItems[item.Key] = item.Value
	}
	put("system", systemItems)

	// CORS
	if cors, err := model.NewCorsConfig().GetConfig(ctx); err == nil {
		put("cors", map[string]string{"allowed_origins": cors.AllowedOrigins})
	}

	// 社区推送（push_key/verify_code 为凭据类）
	if push, err := model.NewCommunityPushConfig().GetConfig(ctx); err == nil {
		put("community", map[string]string{
			"base_url":    push.BaseURL,
			"site_url":    push.SiteURL,
			"push_key":    push.PushKey,
			"verify_code": push.VerifyCode,
		})
	}

	// 主题市场（public_api_base；遗留 market_base_url 不导出）
	if tm, err := model.NewThemeMarketConfig().GetConfig(ctx); err == nil {
		put("theme_market", map[string]string{"public_api_base": tm.PublicAPIBase})
	}

	// 安全参数
	if sec, err := model.NewSecurityConfig().GetConfig(ctx); err == nil {
		put("security", map[string]string{
			"security_enabled":      strconv.FormatBool(sec.SecurityEnabled),
			"blacklist_ttl_minutes": strconv.Itoa(sec.BlacklistTTLMinutes),
			"log_retention_days":    strconv.Itoa(sec.LogRetentionDays),
		})
	}

	// 模块开关
	if modules, err := model.NewModuleConfig().GetConfig(ctx); err == nil {
		items := map[string]string{
			"article":     strconv.FormatBool(modules.ArticleEnabled),
			"music":       strconv.FormatBool(modules.MusicEnabled),
			"video":       strconv.FormatBool(modules.VideoEnabled),
			"travel":      strconv.FormatBool(modules.TravelEnabled),
			"portfolio":   strconv.FormatBool(modules.PortfolioEnabled),
			"equipment":   strconv.FormatBool(modules.EquipmentEnabled),
			"project":     strconv.FormatBool(modules.ProjectEnabled),
			"open_source": strconv.FormatBool(modules.OpenSourceEnabled),
			"recipe":      strconv.FormatBool(modules.RecipeEnabled),
			"book":        strconv.FormatBool(modules.BookEnabled),
			"game":        strconv.FormatBool(modules.GameEnabled),
			"fitness":     strconv.FormatBool(modules.FitnessEnabled),
			"tech_stack":  strconv.FormatBool(modules.TechStackEnabled),
		}
		put("modules", items)
	}

	// 地图（Key 为凭据类）
	if mp, err := model.NewMapConfig().GetConfig(ctx); err == nil {
		put("map", map[string]string{
			"amap_key":           mp.AmapKey,
			"amap_security_code": mp.AmapSecurityCode,
			"google_key":         mp.GoogleKey,
		})
	}

	// 对象存储（access_secret 永不导出）
	storageList, err := model.NewStorageConfig().GetAll(ctx)
	if err == nil {
		items := make(map[string]string)
		for _, s := range storageList {
			prefix := "provider." + s.Provider + "."
			items[prefix+"endpoint"] = s.Endpoint
			items[prefix+"region"] = s.Region
			items[prefix+"bucket"] = s.Bucket
			items[prefix+"custom_domain"] = s.CustomDomain
			items[prefix+"path_prefix"] = s.PathPrefix
		}
		put("storage", items)
	}

	return &res.SystemConfigExportRes{
		Version:         1,
		ExportedAt:      time.Now(),
		SecretsIncluded: includeSecrets,
		Groups:          groups,
	}, nil
}

// ApplyConfigImport 导入配置：仅处理已知分组与已知键，占位符跳过。
// 返回各分组实际写入的键数。storage/map/cloud 的凭据类字段不在此导入。
func (l *SystemConfigLogic) ApplyConfigImport(ctx context.Context, export *res.SystemConfigExportRes) (map[string]int, error) {
	applied := make(map[string]int)
	write := func(group, key, value string) error {
		switch group {
		case "system":
			if err := l.set(ctx, key, value, descriptionOf(key)); err != nil {
				return err
			}
			applied[group]++
		case "cors":
			if key == "allowed_origins" {
				if err := NewCorsConfigLogic().UpdateConfig(ctx, &req.UpdateCorsConfigReq{AllowedOrigins: &value}); err != nil {
					return err
				}
				applied[group]++
			}
		case "community":
			// 仅导入非空的地址类字段(空值跳过,避免清空现有配置);push_key 留空即不覆盖
			if value == "" || (key != "base_url" && key != "site_url") {
				return nil
			}
			pushReq := UpdatePushConfigRequest{BaseURL: value}
			if key == "site_url" {
				pushReq = UpdatePushConfigRequest{SiteURL: value}
			}
			if _, err := NewCommunityPushLogic().UpdatePushConfig(ctx, &pushReq); err != nil {
				return err
			}
			applied[group]++
		case "theme_market":
			if key == "public_api_base" {
				if err := NewThemeMarketConfigLogic().UpdateConfig(ctx, &req.UpdateThemeMarketConfigReq{PublicAPIBase: &value}); err != nil {
					return err
				}
				applied[group]++
			}
		case "security":
			secReq, err := buildSecurityImportReq(key, value)
			if err != nil {
				return nil
			}
			if err := NewSecurityLogic().UpdateConfig(ctx, secReq); err != nil {
				return err
			}
			applied[group]++
		case "modules":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil
			}
			moduleReq := buildModuleImportReq(key, b)
			if moduleReq == nil {
				return nil
			}
			if err := NewModuleConfigLogic().UpdateConfig(ctx, moduleReq); err != nil {
				return err
			}
			applied[group]++
		}
		return nil
	}

	for group, items := range export.Groups {
		keys := make([]string, 0, len(items))
		for k := range items {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := items[k]
			if v == secretPlaceholder {
				continue
			}
			if err := write(group, k, v); err != nil {
				return applied, fmt.Errorf("导入 %s.%s 失败: %w", group, k, err)
			}
		}
	}
	return applied, nil
}

// isSecretField 判定导出脱敏字段。
func isSecretField(group, key string) bool {
	switch group {
	case "community":
		return key == "push_key" || key == "verify_code"
	case "map":
		return key == "amap_key" || key == "amap_security_code" || key == "google_key"
	}
	return false
}

// buildSecurityImportReq 把导出键值转换为安全配置部分更新请求(无法识别的键返回 nil)。
func buildSecurityImportReq(key, value string) (*req.UpdateSecurityConfigReq, error) {
	switch key {
	case "security_enabled":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &req.UpdateSecurityConfigReq{SecurityEnabled: &b}, nil
	case "blacklist_ttl_minutes":
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &req.UpdateSecurityConfigReq{BlacklistTTLMinutes: &n}, nil
	case "log_retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &req.UpdateSecurityConfigReq{LogRetentionDays: &n}, nil
	}
	return nil, fmt.Errorf("未知的安全配置键: %s", key)
}

// moduleEnabledFields 导出键 → 模块开关 DTO 字段映射。
func moduleEnabledFields() map[string]func(*req.UpdateModuleConfigReq, *bool) {
	return map[string]func(*req.UpdateModuleConfigReq, *bool){
		"article":     func(r *req.UpdateModuleConfigReq, v *bool) { r.ArticleEnabled = v },
		"music":       func(r *req.UpdateModuleConfigReq, v *bool) { r.MusicEnabled = v },
		"video":       func(r *req.UpdateModuleConfigReq, v *bool) { r.VideoEnabled = v },
		"travel":      func(r *req.UpdateModuleConfigReq, v *bool) { r.TravelEnabled = v },
		"portfolio":   func(r *req.UpdateModuleConfigReq, v *bool) { r.PortfolioEnabled = v },
		"equipment":   func(r *req.UpdateModuleConfigReq, v *bool) { r.EquipmentEnabled = v },
		"project":     func(r *req.UpdateModuleConfigReq, v *bool) { r.ProjectEnabled = v },
		"open_source": func(r *req.UpdateModuleConfigReq, v *bool) { r.OpenSourceEnabled = v },
		"recipe":      func(r *req.UpdateModuleConfigReq, v *bool) { r.RecipeEnabled = v },
		"book":        func(r *req.UpdateModuleConfigReq, v *bool) { r.BookEnabled = v },
		"game":        func(r *req.UpdateModuleConfigReq, v *bool) { r.GameEnabled = v },
		"fitness":     func(r *req.UpdateModuleConfigReq, v *bool) { r.FitnessEnabled = v },
		"tech_stack":  func(r *req.UpdateModuleConfigReq, v *bool) { r.TechStackEnabled = v },
	}
}

// buildModuleImportReq 把导出键值转换为模块开关部分更新请求(无法识别的键返回 nil)。
func buildModuleImportReq(key string, enabled bool) *req.UpdateModuleConfigReq {
	setters := moduleEnabledFields()
	setter, ok := setters[key]
	if !ok {
		return nil
	}
	r := &req.UpdateModuleConfigReq{}
	setter(r, &enabled)
	return r
}
