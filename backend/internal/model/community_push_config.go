package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommunityPushConfig CMS 推送到 NovaBlog 社区的运行配置，对应 community_push_configs 表（单行模式）。
type CommunityPushConfig struct {
	ID      string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BaseURL string `gorm:"type:varchar(500);not null;default:''"` // 社区 API 地址（空=未配置）
	SiteURL string `gorm:"type:varchar(500);not null;default:''"` // 博客对外地址（拼 source_url）
	PushKey string `gorm:"type:varchar(200);not null;default:''"` // 推送密钥 nbk_...（仅管理员接口可写，读取脱敏）
	CreatedAt  time.Time `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName 指定数据表名称。
func (CommunityPushConfig) TableName() string {
	return "community_push_configs"
}

// CommunityPushConfigModel 推送配置模型操作结构体。
type CommunityPushConfigModel struct {
	db *gorm.DB
}

// NewCommunityPushConfig 创建 CommunityPushConfigModel 实例。
func NewCommunityPushConfig() *CommunityPushConfigModel {
	return &CommunityPushConfigModel{db: DB}
}

// GetConfig 获取推送配置，不存在则插入默认空配置并返回。
func (m *CommunityPushConfigModel) GetConfig(ctx context.Context) (*CommunityPushConfig, error) {
	var config CommunityPushConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err == nil {
		return &config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	config = CommunityPushConfig{ID: uuid.New().String()}
	if err := m.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfig 全量更新推送配置。
func (m *CommunityPushConfigModel) UpdateConfig(ctx context.Context, config *CommunityPushConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}
