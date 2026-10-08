package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MapConfig 地图服务配置模型，对应 map_configs 数据表（单行模式）。
// amap_security_code 落库前经 AES-GCM 加密（crypto.secret_key），读取时解密。
type MapConfig struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AmapKey          string    `gorm:"type:varchar(200);not null;default:''" json:"amap_key"`
	AmapSecurityCode string    `gorm:"type:text;not null;default:''" json:"amap_security_code"`
	GoogleKey        string    `gorm:"type:varchar(200);not null;default:''" json:"google_key"`
	CreatedAt        time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName 指定数据表名称。
func (MapConfig) TableName() string {
	return "map_configs"
}

// MapConfigModel 地图服务配置模型操作结构体。
type MapConfigModel struct {
	db *gorm.DB
}

// NewMapConfig 创建 MapConfigModel 实例。
func NewMapConfig() *MapConfigModel {
	return &MapConfigModel{db: DB}
}

// GetConfig 获取地图服务配置，如果不存在则插入默认配置并返回。
func (m *MapConfigModel) GetConfig(ctx context.Context) (*MapConfig, error) {
	var config MapConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err == nil {
		return &config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// 不存在则创建默认配置（Key 为空表示未配置，选点自动降级 OSM 免 Key 模式）
	config = MapConfig{
		ID: uuid.New().String(),
	}
	if createErr := m.db.WithContext(ctx).Create(&config).Error; createErr != nil {
		return nil, createErr
	}
	return &config, nil
}

// UpdateConfig 更新地图服务配置（全量更新）。
func (m *MapConfigModel) UpdateConfig(ctx context.Context, config *MapConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}
