package bootstrap

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/logic"
	"novablog/internal/middleware"
	"novablog/internal/model"
	"novablog/utils/response"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用全局结构体，持有所有核心依赖。
type App struct {
	Config *viper.Viper  // 配置实例
	Logger *zap.Logger   // 日志实例
	DB     *gorm.DB      // 数据库实例
	Redis  *redis.Client // Redis 实例
}

// NewApp 创建并初始化应用实例。
func NewApp(cfgPath string) (*App, error) {
	// 加载配置
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 初始化日志
	logger, err := InitLogger(&LogConfig{
		Level:      cfg.GetString("log.level"),
		FilePath:   cfg.GetString("log.file_path"),
		MaxSize:    cfg.GetInt("log.max_size"),
		MaxBackups: cfg.GetInt("log.max_backups"),
		MaxAge:     cfg.GetInt("log.max_age"),
	})
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	// 初始化数据库
	db, err := InitDB(&DBConfig{
		Host:         cfg.GetString("postgres.host"),
		Port:         cfg.GetInt("postgres.port"),
		User:         cfg.GetString("postgres.user"),
		Password:     cfg.GetString("postgres.password"),
		DBName:       cfg.GetString("postgres.dbname"),
		SSLMode:      cfg.GetString("postgres.sslmode"),
		MaxOpenConns: cfg.GetInt("postgres.max_open_conns"),
		MaxIdleConns: cfg.GetInt("postgres.max_idle_conns"),
	})
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化 Redis
	rdb, err := InitRedis(&RedisConfig{
		Host:     cfg.GetString("redis.host"),
		Port:     cfg.GetInt("redis.port"),
		Password: cfg.GetString("redis.password"),
		DB:       cfg.GetInt("redis.db"),
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 Redis 失败: %w", err)
	}

	// 注入全局变量
	logic.AccessSecret = cfg.GetString("jwt.access_secret")
	logic.RefreshSecret = cfg.GetString("jwt.refresh_secret")
	logic.AccessExpire = cfg.GetDuration("jwt.access_expire")
	logic.RefreshExpire = cfg.GetDuration("jwt.refresh_expire")
	// 官方部署安装码：非空时首装初始化必须携带（防恶意抢先初始化），自部署留空不启用
	logic.InstallInitCode = cfg.GetString("install.init_code")
	logic.AuthLogger = logger
	logic.SetupLogger = logger
	logic.MusicLogger = logger
	logic.CommentLogger = logger
	logic.MediaLogger = logger

	// 注入主题模块配置（官方站点地址统一由 official_site_settings 托管，见下方解析）
	logic.SetThemeSettings(&logic.ThemeSettings{
		DataDir:       cfg.GetString("themes.data_dir"),
		FrontendDir:   cfg.GetString("themes.frontend_dir"),
		GithubToken:   cfg.GetString("themes.github_token"),
		PublicAPIBase: cfg.GetString("themes.public_api_base"),
		ProxyURL:      cfg.GetString("themes.proxy_url"),
		MaxArtifactMB: cfg.GetInt("themes.max_artifact_mb"),
	})
	logic.ThemeLogger = logger

	// 注入开源作品模块配置：与主题模块共用 GitHub 访问配置（token 提升限流额度、代理走探测降级）
	logic.SetOpenSourceSettings(&logic.OpenSourceSettings{
		GithubToken: cfg.GetString("themes.github_token"),
		ProxyURL:    cfg.GetString("themes.proxy_url"),
	})

	// 注入图片搜索模块配置（管理端在线搜图标/封面，代理不可用自动降级直连）
	logic.SetImageSearchSettings(&logic.ImageSearchSettings{
		ProxyURL: cfg.GetString("image_search.proxy_url"),
	})

	// 注入社区推送模块配置（官方站点地址统一由 official_site_settings 托管，见下方解析）
	logic.SetCommunityPushSettings(&logic.CommunityPushSettings{
		SiteURL: cfg.GetString("community.site_url"),
	})
	logic.CommunityPushLogger = logger

	// 主题模块运行配置：DB 持久化值（后台修改）优先于 config.yaml 出厂值；
	// 旧版主题市场地址仅作官方站点地址的存量兜底，不再写入
	var legacyThemeBase string
	if themeConfig, err := model.NewThemeMarketConfig().GetConfig(context.Background()); err == nil {
		logic.SetThemePublicAPIBase(themeConfig.PublicAPIBase)
		legacyThemeBase = themeConfig.MarketBaseURL
	} else {
		logger.Warn("读取主题运行配置失败，使用 config.yaml 出厂默认值", zap.Error(err))
	}

	// 社区推送配置：DB 持久化值优先（后台修改），config.yaml 出厂值兜底
	officialBase := ""
	if pushConfig, err := model.NewCommunityPushConfig().GetConfig(context.Background()); err == nil {
		officialBase = pushConfig.BaseURL
		logic.SetCommunitySiteURL(pushConfig.SiteURL)
		logic.SetCommunityPushKey(pushConfig.PushKey)
	} else {
		logger.Warn("读取社区推送配置失败，使用 config.yaml 出厂默认值", zap.Error(err))
	}

	// 官方站点地址统一解析（社区推送与主题市场共用，后台「社区推送配置」页为唯一维护入口）：
	// 社区推送 DB 权威值 → community.base_url 出厂值 → 旧版主题市场 DB 值 → 旧版 themes 出厂值
	if officialBase == "" {
		officialBase = cfg.GetString("community.base_url")
	}
	if officialBase == "" {
		officialBase = legacyThemeBase
	}
	if officialBase == "" {
		officialBase = cfg.GetString("themes.market_base_url")
	}
	logic.SetOfficialSiteBaseURL(officialBase)

	// 注入跨域配置：DB 是唯一权威（后台保存的值不被任何来源覆盖）；
	// config.yaml 仅在 DB 无有效数据时作为初始值播种，环境变量不再参与
	corsLogic := logic.NewCorsConfigLogic()
	if err := corsLogic.InitFromFallback(context.Background(), cfg.GetString("cors.allowed_origins")); err != nil {
		logger.Warn("跨域配置引导失败，中间件使用默认白名单", zap.Error(err))
	}

	middleware.AuthLogger = logger
	middleware.Logger = logger
	response.ErrorLogger = logger

	app := &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Redis:  rdb,
	}

	logger.Info("应用初始化完成",
		zap.String("app", cfg.GetString("app.name")),
		zap.String("env", cfg.GetString("app.env")),
	)

	return app, nil
}

// loadConfig 加载配置文件。
func loadConfig(cfgPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将 duration 字符串（如 "2h"）转换为 time.Duration 并在 viper 中可用
	v.Set("jwt.access_expire", parseDurationSafe(v.GetString("jwt.access_expire"), 2*time.Hour))
	v.Set("jwt.refresh_expire", parseDurationSafe(v.GetString("jwt.refresh_expire"), 168*time.Hour))

	return v, nil
}

// parseDurationSafe 安全解析 duration 字符串，失败时返回默认值。
func parseDurationSafe(s string, defaultDur time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultDur
	}
	return d
}
