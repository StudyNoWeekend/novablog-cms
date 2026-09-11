package bootstrap

import (
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
