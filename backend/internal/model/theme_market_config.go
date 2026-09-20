package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ThemeMarketConfig 主题模块运行配置模型，对应 theme_market_configs 数据表（单行模式）。
type ThemeMarketConfig struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MarketBaseURL string    `gorm:"type:varchar(500);not null;default:''"`
	PublicAPIBase string    `gorm:"type:varchar(500);not null;default:''"`
	CreatedAt     time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (ThemeMarketConfig) TableName() string {
	return "theme_market_configs"
}

// ThemeMarketConfigModel 官方主题市场配置模型操作结构体。
type ThemeMarketConfigModel struct {
	db *gorm.DB
}

// NewThemeMarketConfig 创建 ThemeMarketConfigModel 实例。
func NewThemeMarketConfig() *ThemeMarketConfigModel {
	return &ThemeMarketConfigModel{db: DB}
}

// GetConfig 获取官方市场配置，如果不存在则插入默认配置并返回。
func (m *ThemeMarketConfigModel) GetConfig(ctx context.Context) (*ThemeMarketConfig, error) {
	var config ThemeMarketConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err == nil {
		return &config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// 不存在则创建默认配置（空地址 = 未自定义，使用 config.yaml 出厂值）
	config = ThemeMarketConfig{
		ID: uuid.New().String(),
	}
	if createErr := m.db.WithContext(ctx).Create(&config).Error; createErr != nil {
		return nil, createErr
	}
	return &config, nil
}

// UpdateConfig 更新官方市场配置（全量更新）。
func (m *ThemeMarketConfigModel) UpdateConfig(ctx context.Context, config *ThemeMarketConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}
