package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecurityConfig 安全配置模型，对应 security_configs 数据表。
type SecurityConfig struct {
	ID                  string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // UUID 主键
	GetMaxTokens        int       `gorm:"not null;default:20"`                            // GET 请求最大令牌数
	GetWindowSeconds    int       `gorm:"not null;default:60"`                            // GET 请求时间窗口（秒）
	PostMaxTokens       int       `gorm:"not null;default:5"`                             // POST 请求最大令牌数
	PostWindowSeconds   int       `gorm:"not null;default:60"`                            // POST 请求时间窗口（秒）
	ViewMaxTokens       int       `gorm:"not null;default:10"`                            // 浏览相关接口最大令牌数
	ViewWindowSeconds   int       `gorm:"not null;default:60"`                            // 浏览相关接口时间窗口（秒）
	LikeMaxTokens       int       `gorm:"not null;default:10"`                            // 点赞相关接口最大令牌数
	LikeWindowSeconds   int       `gorm:"not null;default:60"`                            // 点赞相关接口时间窗口（秒）
	BlacklistThreshold  int       `gorm:"not null;default:5"`                             // 触发黑名单的违规次数阈值
	BlacklistTTLMinutes int       `gorm:"not null;default:60"`                            // 黑名单封禁时长（分钟）
	CreatedAt           time.Time `gorm:"type:timestamptz;autoCreateTime"`                // 创建时间
	UpdatedAt           time.Time `gorm:"type:timestamptz;autoUpdateTime"`                // 更新时间
}

// TableName 指定数据表名称。
func (SecurityConfig) TableName() string {
	return "security_configs"
}

// SecurityConfigModel 安全配置模型操作结构体。
type SecurityConfigModel struct {
	db *gorm.DB
}

// NewSecurityConfig 创建 SecurityConfigModel 实例。
func NewSecurityConfig() *SecurityConfigModel {
	return &SecurityConfigModel{db: DB}
}

// GetConfig 获取安全配置，如果不存在则插入默认配置并返回。
func (m *SecurityConfigModel) GetConfig(ctx context.Context) (*SecurityConfig, error) {
	var config SecurityConfig
	err := m.db.WithContext(ctx).First(&config).Error
	if err == nil {
		return &config, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// 不存在则创建默认配置
	config = SecurityConfig{
		ID:                  uuid.New().String(),
		GetMaxTokens:        20,
		GetWindowSeconds:    60,
		PostMaxTokens:       5,
		PostWindowSeconds:   60,
		ViewMaxTokens:       10,
		ViewWindowSeconds:   60,
		LikeMaxTokens:       10,
		LikeWindowSeconds:   60,
		BlacklistThreshold:  5,
		BlacklistTTLMinutes: 60,
	}
	if createErr := m.db.WithContext(ctx).Create(&config).Error; createErr != nil {
		return nil, createErr
	}
	return &config, nil
}

// UpdateConfig 更新安全配置。
func (m *SecurityConfigModel) UpdateConfig(ctx context.Context, config *SecurityConfig) error {
	return m.db.WithContext(ctx).Save(config).Error
}

// IPBlacklistRecord IP 黑名单记录模型，对应 ip_blacklist 数据表。
type IPBlacklistRecord struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // UUID 主键
	IPAddress string    `gorm:"type:varchar(50);index;not null"`                // IP 地址
	Reason    string    `gorm:"type:varchar(200)"`                              // 封禁原因
	BannedAt  time.Time `gorm:"type:timestamptz"`                               // 封禁时间
	ExpiresAt time.Time `gorm:"type:timestamptz"`                               // 过期时间
	IsActive  bool      `gorm:"default:true"`                                   // 是否生效
	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime"`                // 创建时间
}

// TableName 指定数据表名称。
func (IPBlacklistRecord) TableName() string {
	return "ip_blacklist"
}

// IPBlacklistRecordModel IP 黑名单记录模型操作结构体。
type IPBlacklistRecordModel struct {
	db *gorm.DB
}

// NewIPBlacklistRecord 创建 IPBlacklistRecordModel 实例。
func NewIPBlacklistRecord() *IPBlacklistRecordModel {
	return &IPBlacklistRecordModel{db: DB}
}

// Create 创建 IP 黑名单记录。
func (m *IPBlacklistRecordModel) Create(ctx context.Context, record *IPBlacklistRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	return m.db.WithContext(ctx).Create(record).Error
}

// GetActiveByIP 根据 IP 地址查询生效的黑名单记录。
func (m *IPBlacklistRecordModel) GetActiveByIP(ctx context.Context, ip string) (*IPBlacklistRecord, error) {
	var record IPBlacklistRecord
	err := m.db.WithContext(ctx).
		Where("ip_address = ? AND is_active = true", ip).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetList 分页查询 IP 黑名单记录，按封禁时间倒序排列。
func (m *IPBlacklistRecordModel) GetList(ctx context.Context, page, pageSize int) ([]IPBlacklistRecord, int64, error) {
	var records []IPBlacklistRecord
	var total int64

	db := m.db.WithContext(ctx).Model(&IPBlacklistRecord{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("banned_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// Delete 根据 IP 地址软删除黑名单记录（将 is_active 置为 false）。
func (m *IPBlacklistRecordModel) Delete(ctx context.Context, ip string) error {
	return m.db.WithContext(ctx).
		Model(&IPBlacklistRecord{}).
		Where("ip_address = ? AND is_active = true", ip).
		Update("is_active", false).Error
}

// GetActiveCount 查询当前生效的黑名单记录数量。
func (m *IPBlacklistRecordModel) GetActiveCount(ctx context.Context) (int64, error) {
	var count int64
	err := m.db.WithContext(ctx).
		Model(&IPBlacklistRecord{}).
		Where("is_active = true").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
