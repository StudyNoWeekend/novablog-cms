package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"novablog/bootstrap"
	"novablog/internal/model"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// 种子命令入口。
// 默认从 backend 目录运行：
//
//	go run ./cmd/seed
//
// 可通过参数指定配置文件与 SQL 脚本路径：
//
//	go run ./cmd/seed config/config.yaml scripts/seed_photographer_data.sql
func main() {
	cfgPath := "config/config.yaml"
	sqlPath := "scripts/seed_photographer_data.sql"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		sqlPath = os.Args[2]
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	db, err := bootstrap.InitDB(&bootstrap.DBConfig{
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
		fmt.Fprintf(os.Stderr, "初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	model.BaseURL = cfg.GetString("upload.base_url")

	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取种子脚本失败: %v\n", err)
		os.Exit(1)
	}

	// 去掉 SQL 文件中的显式事务边界，改由 GORM 事务统一控制。
	sql := strings.ReplaceAll(string(sqlBytes), "BEGIN;", "")
	sql = strings.ReplaceAll(sql, "COMMIT;", "")

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 禁用 prepared statement，避免 pgx 对多语句执行的拦截。
		tx = tx.Session(&gorm.Session{PrepareStmt: false})
		return tx.WithContext(context.Background()).Exec(sql).Error
	}); err != nil {
		fmt.Fprintf(os.Stderr, "执行种子脚本失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("摄影师示例数据导入成功")
}

func loadConfig(cfgPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
