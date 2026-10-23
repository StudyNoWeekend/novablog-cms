package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 云盘账号状态。
const (
	CloudAccountStatusActive   = "active"
	CloudAccountStatusExpired  = "expired"
	CloudAccountStatusRevoked  = "revoked"
)

// 云盘 provider。
const (
	CloudProviderWebDAV = "webdav"
	CloudProviderBaidu  = "baidu"
	CloudProviderAliyun = "aliyun"
)

// CloudDriveAccount 云盘授权账号模型，对应 cloud_drive_accounts 数据表。
// OAuth 盘（baidu/aliyun）存 access/refresh token；WebDAV 复用 AccessToken 字段存
// AES 加密后的密码，BaseURL/Username 记录服务器地址与登录名（一个服务可绑多个账号）。
type CloudDriveAccount struct {
	ID           string         `gorm:"type:uuid;primaryKey"`
	Provider     string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_provider_open_user"`
	OpenUserID   string         `gorm:"column:open_user_id;type:varchar(128);not null;uniqueIndex:idx_provider_open_user"` // OAuth=盘方用户 ID；WebDAV=base_url|username
	DisplayName  string         `gorm:"type:varchar(128)"`
	AccessToken  string         `gorm:"type:text"` // OAuth access token / WebDAV 密码（均 AES 加密落库）
	RefreshToken string         `gorm:"type:text"` // AES 加密，WebDAV 为空
	ExpiresAt    *time.Time     `gorm:"column:expires_at;type:timestamptz"`
	BaseURL      string         `gorm:"type:varchar(255)"` // WebDAV 服务器地址
	Username     string         `gorm:"type:varchar(128)"` // WebDAV 用户名
	DriveID      string         `gorm:"type:varchar(64)"`  // 阿里云盘 default_drive_id
	Status       string         `gorm:"type:varchar(20);not null;default:'active'"`
	QuotaTotal   int64          `gorm:"type:bigint;default:0"`
	QuotaUsed    int64          `gorm:"type:bigint;default:0"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName 指定数据表名称。
func (CloudDriveAccount) TableName() string {
	return "cloud_drive_accounts"
}

// CloudDriveAccountModel 云盘授权账号模型操作结构体。
type CloudDriveAccountModel struct {
	db *gorm.DB
}

// NewCloudDriveAccount 创建 CloudDriveAccountModel 实例。
func NewCloudDriveAccount() *CloudDriveAccountModel {
	return &CloudDriveAccountModel{db: DB}
}

// GetAll 返回全部账号（未软删），按 created_at 排序。
func (m *CloudDriveAccountModel) GetAll(ctx context.Context) ([]CloudDriveAccount, error) {
	var accounts []CloudDriveAccount
	err := m.db.WithContext(ctx).Order("created_at").Find(&accounts).Error
	return accounts, err
}

// GetByID 按 ID 查询账号（未软删）。
func (m *CloudDriveAccountModel) GetByID(ctx context.Context, id string) (*CloudDriveAccount, error) {
	var account CloudDriveAccount
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByProviderAndOpenUser 按 provider+open_user_id 查询账号（含软删记录，用于重复绑定时复活）。
func (m *CloudDriveAccountModel) GetByProviderAndOpenUser(ctx context.Context, provider, openUserID string) (*CloudDriveAccount, error) {
	var account CloudDriveAccount
	err := m.db.WithContext(ctx).Unscoped().
		Where("provider = ? AND open_user_id = ?", provider, openUserID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Create 创建账号记录。
func (m *CloudDriveAccountModel) Create(ctx context.Context, account *CloudDriveAccount) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	return m.db.WithContext(ctx).Create(account).Error
}

// Update 更新账号记录。
func (m *CloudDriveAccountModel) Update(ctx context.Context, account *CloudDriveAccount) error {
	return m.db.WithContext(ctx).Save(account).Error
}

// SoftDelete 解绑账号（软删除）。
func (m *CloudDriveAccountModel) SoftDelete(ctx context.Context, id string) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Delete(&CloudDriveAccount{}).Error
}
