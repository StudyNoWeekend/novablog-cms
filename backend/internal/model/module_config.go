package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModuleConfig 模块开关配置模型，对应 module_configs 数据表（单行模式）。
type ModuleConfig struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ArticleEnabled   bool      `gorm:"not null;default:true"`
	MediaEnabled     bool      `gorm:"not null;default:true"`
	MusicEnabled     bool      `gorm:"not null;default:true"`
	VideoEnabled     bool      `gorm:"not null;default:true"`
	TravelEnabled    bool      `gorm:"not null;default:true"`
	PortfolioEnabled bool      `gorm:"not null;default:true"`
	EquipmentEnabled bool      `gorm:"not null;default:true"`
	ProjectEnabled   bool      `gorm:"not null;default:true"`
	CreatedAt        time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (ModuleConfig) TableName() string {
	return "module_configs"
}

// ModuleConfigModel 模块开关配置模型操作结构体。
type ModuleConfigModel struct {
	db *gorm.DB
}

// NewModuleConfig 创建 ModuleConfigModel 实例。
func NewModuleConfig() *ModuleConfigModel {
	return &ModuleConfigModel{db: DB}
}

// GetConfig 获取模块开关配置，如果不存在则插入默认配置并返回。
func (m *ModuleConfigModel) GetConfig(ctx context.Context) (*ModuleConfig, error) {
	var config ModuleConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err == nil {
		return &config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// 不存在则创建默认配置
	config = ModuleConfig{
		ID:               uuid.New().String(),
		ArticleEnabled:   true,
		MediaEnabled:     true,
		MusicEnabled:     true,
		VideoEnabled:     true,
		TravelEnabled:    true,
		PortfolioEnabled: true,
		EquipmentEnabled: true,
		ProjectEnabled:   true,
	}
	if createErr := m.db.WithContext(ctx).Create(&config).Error; createErr != nil {
		return nil, createErr
	}
	return &config, nil
}

// UpdateConfig 更新模块开关配置（全量更新）。
func (m *ModuleConfigModel) UpdateConfig(ctx context.Context, config *ModuleConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}
