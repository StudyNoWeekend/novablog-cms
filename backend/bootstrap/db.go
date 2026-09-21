package bootstrap

import (
	"database/sql"
	"fmt"
	"time"

	"novablog/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// DBConfig 数据库配置参数。
type DBConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

// InitDB 初始化 PostgreSQL 数据库连接，自动迁移模型。
// 如果数据库不存在，会自动创建数据库，然后迁移所有模型表。
func InitDB(cfg *DBConfig) (*gorm.DB, error) {
	// 1. 先连接默认数据库 postgres，检查目标数据库是否存在
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, fmt.Errorf("确保数据库存在失败: %w", err)
	}

	// 2. 连接目标数据库
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 3. 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 4. 自动迁移所有模型
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	// 5. 数据一致性回填（幂等）：用真实评论数校正 comment_count/review_count 计数
	backfillCounters(db)

	// 6. 设置全局 DB
	model.DB = db

	return db, nil
}

// ensureDatabaseExists 检查数据库是否存在，不存在则创建。
func ensureDatabaseExists(cfg *DBConfig) error {
	// 连接默认数据库 postgres
	defaultDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return fmt.Errorf("连接默认数据库失败: %w", err)
	}
	defer db.Close()

	// 检查数据库是否存在（使用参数化查询防止 SQL 注入）
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("查询数据库是否存在失败: %w", err)
	}

	// 如果不存在则创建
	if !exists {
		// PostgreSQL 中创建数据库需要使用标识符引号，防止特殊字符问题
		createSQL := fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.DBName)
		if _, err := db.Exec(createSQL); err != nil {
			return fmt.Errorf("创建数据库失败: %w", err)
		}
	}

	return nil
}

// autoMigrate 自动迁移所有模型表。
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Blogger{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleTag{},
		&model.Media{},
		&model.MediaPreset{},
		&model.StorageConfig{},
		&model.StorageMigrationTask{},
		&model.StorageMigrationItem{},
		&model.TravelGuide{},
		&model.Song{},
		&model.Comment{},
		&model.VideoWork{},
		&model.VideoPlatformLink{},
		&model.Portfolio{},
		&model.PortfolioItem{},
		&model.SecurityConfig{},
		&model.IPBlacklistRecord{},
		&model.IPBlacklist{},
		&model.AccessLog{},
		&model.ModuleConfig{},
		&model.Theme{},
		&model.PhotoEquipment{},
		&model.Project{},
		&model.OpenSourceWork{},
		&model.CorsConfig{},
		&model.ThirdPartyPlaylist{},
		&model.ThemeMarketConfig{},
	)
}

// backfillCounters 用真实数据校正内容计数（幂等，每次启动执行一次）：
//   - articles.comment_count ← comments 表中 target_type='article' 的未删除评论数
//   - travel_guides.review_count ← comments 表中 target_type='travel_guide' 的未删除评论数
//
// 修复历史遗留：评论计数器从未被维护导致文章管理/热门排行始终显示 0。
// 仅校正有评论但计数明显偏小（相差 >= 1）的内容，其余保持不变。
func backfillCounters(db *gorm.DB) {
	// 文章评论数回填
	if err := db.Exec(`
		UPDATE articles a
		SET comment_count = t.real_count
		FROM (
			SELECT c.target_id, COUNT(*) AS real_count
			FROM comments c
			WHERE c.target_type = 'article' AND c.deleted_at IS NULL
			GROUP BY c.target_id
		) t
		WHERE a.id = t.target_id
		  AND t.real_count > a.comment_count
	`).Error; err != nil {
		// 回填失败不影响启动
		return
	}

	// 旅行攻略评论数回填
	_ = db.Exec(`
		UPDATE travel_guides tg
		SET review_count = t.real_count
		FROM (
			SELECT c.target_id, COUNT(*) AS real_count
			FROM comments c
			WHERE c.target_type = 'travel_guide' AND c.deleted_at IS NULL
			GROUP BY c.target_id
		) t
		WHERE tg.id = t.target_id
		  AND t.real_count > tg.review_count
	`).Error
}
