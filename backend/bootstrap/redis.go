package bootstrap

import (
	"context"
	"fmt"
	"time"

	"cus-cms/internal/cache"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis 配置参数。
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// InitRedis 初始化 Redis 客户端连接。
func InitRedis(cfg *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Ping 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	// 设置全局 RedisClient
	cache.RedisClient = client

	return client, nil
}
