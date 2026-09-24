package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CorsConfig 跨域配置模型，对应 cors_configs 数据表（单行模式）。
type CorsConfig struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AllowedOrigins string    `gorm:"type:text;not null;default:'http://localhost:5173,http://localhost:5174'"`
	CreatedAt      time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (CorsConfig) TableName() string {
	return "cors_configs"
}

// CorsConfigModel 跨域配置模型操作结构体。
type CorsConfigModel struct {
	db *gorm.DB
}

// NewCorsConfig 创建 CorsConfigModel 实例。
func NewCorsConfig() *CorsConfigModel {
	return &CorsConfigModel{db: DB}
}

// GetConfig 获取跨域配置记录（纯读取，不自动创建；无记录时返回 gorm.ErrRecordNotFound）。
func (m *CorsConfigModel) GetConfig(ctx context.Context) (*CorsConfig, error) {
	var config CorsConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateConfig 创建跨域配置记录（DB 无数据时由启动引导播种或首次保存调用）。
func (m *CorsConfigModel) CreateConfig(ctx context.Context, origins string) (*CorsConfig, error) {
	config := CorsConfig{
		ID:             uuid.New().String(),
		AllowedOrigins: origins,
	}
	if err := m.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfig 更新跨域配置（全量更新）。
func (m *CorsConfigModel) UpdateConfig(ctx context.Context, config *CorsConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}
