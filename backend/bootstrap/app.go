package bootstrap

import (
	"context"
	"fmt"
	"time"

	"novablog/internal/logic"
	"novablog/internal/middleware"
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
	logic.AuthLogger = logger
	logic.SetupLogger = logger
	logic.MusicLogger = logger

	// 注入主题模块配置
	logic.SetThemeSettings(&logic.ThemeSettings{
		DataDir:       cfg.GetString("themes.data_dir"),
		FrontendDir:   cfg.GetString("themes.frontend_dir"),
		MarketBaseURL: cfg.GetString("themes.market_base_url"),
		GithubToken:   cfg.GetString("themes.github_token"),
		PublicAPIBase: cfg.GetString("themes.public_api_base"),
		MaxArtifactMB: cfg.GetInt("themes.max_artifact_mb"),
	})
	logic.ThemeLogger = logger

	// 主题模块运行配置：DB 持久化值（后台修改）优先于 config.yaml 出厂值
	if marketConfig, err := logic.NewThemeMarketConfigLogic().GetConfig(context.Background()); err == nil {
		logic.SetThemeMarketBaseURL(marketConfig.MarketBaseURL)
		logic.SetThemePublicAPIBase(marketConfig.PublicAPIBase)
	} else {
		logger.Warn("读取主题运行配置失败，使用 config.yaml 出厂默认值", zap.Error(err))
	}

	// 注入跨域配置：优先级 config.yaml → DB 持久化值 → env 变量（中间件内处理）
	corsLogic := logic.NewCorsConfigLogic()
	// 1. 从 config.yaml 读取作为初始值
	cfgCorsOrigin := cfg.GetString("cors.allowed_origins")
	if cfgCorsOrigin != "" {
		middleware.SetAllowedOrigins(cfgCorsOrigin)
	}
	// 2. 尝试从 DB 读取持久化配置——仅当 DB 值不是默认值时覆盖（用户已在后台页面自定义）
	if corsConfig, err := corsLogic.GetConfig(context.Background()); err == nil && corsConfig.AllowedOrigins != "http://localhost:5173,http://localhost:5174" {
		middleware.SetAllowedOrigins(corsConfig.AllowedOrigins)
	}
	// 3. 环境变量 CORS_ALLOWED_ORIGINS 优先级最高，已在中间件内部处理

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
