package logic

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"
	"novablog/utils/crypto"
)

// StorageLogic 存储配置业务逻辑。
type StorageLogic struct {
	model          *model.StorageConfigModel
	manager        *storage.Manager
	cryptoKey      string
	migrationModel *model.StorageMigrationModel // 用于检查迁移锁
	mediaModel     *model.MediaModel            // 用于检查素材迁移状态
}

// NewStorageLogic 创建 StorageLogic 实例。
func NewStorageLogic(manager *storage.Manager, cryptoKey string, migrationModel *model.StorageMigrationModel) *StorageLogic {
	return &StorageLogic{
		model:          model.NewStorageConfig(),
		manager:        manager,
		cryptoKey:      cryptoKey,
		migrationModel: migrationModel,
		mediaModel:     model.NewMedia(),
	}
}

// GetAllConfigs 查询所有存储配置，access_secret 返回 "******" 脱敏。
func (l *StorageLogic) GetAllConfigs(ctx context.Context) ([]res.StorageConfigRes, error) {
	configs, err := l.model.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询存储配置列表失败: %w", err)
	}

	result := make([]res.StorageConfigRes, 0, len(configs))
	for _, c := range configs {
		result = append(result, res.StorageConfigRes{
			ID:           c.ID,
			Provider:     c.Provider,
			Endpoint:     c.Endpoint,
			Region:       c.Region,
			Bucket:       c.Bucket,
			AccessKey:    c.AccessKey,
			AccessSecret: "******",
			PathPrefix:   c.PathPrefix,
			CustomDomain: c.CustomDomain,
			Extra:        c.Extra,
			IsActive:     c.IsActive,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		})
	}
	return result, nil
}

// UpsertConfig 加密 access_secret 后创建或更新存储配置。
// local 存储不需要加密，access_secret 存为空字符串。
// 编辑时 access_secret 留空表示不修改；若更新的配置为活跃配置，保存后自动热重载。
func (l *StorageLogic) UpsertConfig(ctx context.Context, r *req.StorageConfigReq) error {
	// 平台类型变更：old_provider 非空且与 provider 不同
	if r.OldProvider != "" && r.OldProvider != r.Provider {
		// 校验新 provider 无冲突
		_, err := l.model.GetByProvider(ctx, r.Provider)
		if err == nil {
			return fmt.Errorf("该存储平台已有配置，请先删除")
		}

		// 获取旧配置，记录是否激活
		oldConfig, err := l.model.GetByProvider(ctx, r.OldProvider)
		if err != nil {
			return fmt.Errorf("原存储配置不存在: %w", err)
		}
		wasActive := oldConfig.IsActive

		// 若旧配置激活，先取消所有激活，再删除旧配置
		if wasActive {
			// 先将所有配置置为非激活
			if err := l.model.DeactivateAll(ctx); err != nil {
				return fmt.Errorf("取消激活失败: %w", err)
			}
		}
		// 删除旧配置（此时已非激活，Delete 可成功）
		if err := l.model.Delete(ctx, r.OldProvider); err != nil {
			return fmt.Errorf("删除旧配置失败: %w", err)
		}

		// 加密 access_secret 并创建新配置
		accessSecret := ""
		if r.Provider != "local" && r.AccessSecret != "" {
			encrypted, err := crypto.Encrypt(r.AccessSecret, l.cryptoKey)
			if err != nil {
				return fmt.Errorf("加密 access_secret 失败: %w", err)
			}
			accessSecret = encrypted
		}

		// 云存储校验必填字段
		if r.Provider != "local" {
			if r.Endpoint == "" {
				return fmt.Errorf("endpoint 不能为空")
			}
			if r.AccessKey == "" {
				return fmt.Errorf("access_key 不能为空")
			}
			if r.AccessSecret == "" {
				return fmt.Errorf("access_secret 不能为空")
			}
		}

		config := &model.StorageConfig{
			Provider:     r.Provider,
			Endpoint:     r.Endpoint,
			Region:       r.Region,
			Bucket:       r.Bucket,
			AccessKey:    r.AccessKey,
			AccessSecret: accessSecret,
			PathPrefix:   r.PathPrefix,
			CustomDomain: r.CustomDomain,
			Extra:        r.Extra,
			IsActive:     wasActive,
		}

		if err := l.model.Create(ctx, config); err != nil {
			return fmt.Errorf("创建新存储配置失败: %w", err)
		}

		// 若原配置为激活状态，热重载存储管理器
		if wasActive {
			if err := l.manager.Reload(ctx); err != nil {
				return fmt.Errorf("配置已保存，但热重载失败: %w", err)
			}
		}

		return nil
	}

	// 原有逻辑（old_provider 为空或等于 provider）
	existing, err := l.model.GetByProvider(ctx, r.Provider)
	isEditing := err == nil

	// 云存储校验必填字段
	if r.Provider != "local" {
		if r.Endpoint == "" {
			return fmt.Errorf("endpoint 不能为空")
		}
		if r.AccessKey == "" {
			return fmt.Errorf("access_key 不能为空")
		}
		// access_secret 仅在新建时必填，编辑时留空表示不修改
		if !isEditing && r.AccessSecret == "" {
			return fmt.Errorf("access_secret 不能为空")
		}
	}

	// 加密 access_secret（local 跳过；编辑时留空则不更新）
	accessSecret := ""
	if r.Provider != "local" && r.AccessSecret != "" {
		encrypted, err := crypto.Encrypt(r.AccessSecret, l.cryptoKey)
		if err != nil {
			return fmt.Errorf("加密 access_secret 失败: %w", err)
		}
		accessSecret = encrypted
	}

	config := &model.StorageConfig{
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Region:       r.Region,
		Bucket:       r.Bucket,
		AccessKey:    r.AccessKey,
		AccessSecret: accessSecret,
		PathPrefix:   r.PathPrefix,
		CustomDomain: r.CustomDomain,
		Extra:        r.Extra,
	}

	if err := l.model.Upsert(ctx, config); err != nil {
		return fmt.Errorf("保存存储配置失败: %w", err)
	}

	// 若更新的配置为活跃配置，自动热重载使新密钥立即生效
	if isEditing && existing.IsActive {
		if err := l.manager.Reload(ctx); err != nil {
			return fmt.Errorf("配置已保存，但热重载失败: %w", err)
		}
	}

	return nil
}

// DeleteConfig 删除指定 provider 的存储配置（不能删除激活中的配置）。
func (l *StorageLogic) DeleteConfig(ctx context.Context, provider string) error {
	if err := l.model.Delete(ctx, provider); err != nil {
		return fmt.Errorf("删除存储配置失败: %w", err)
	}
	return nil
}

// ActivateConfig 激活指定 provider 的存储配置。
// 迁移进行中禁止切换，所有素材必须已迁移到目标平台，切换后热重载存储管理器。
func (l *StorageLogic) ActivateConfig(ctx context.Context, provider string) error {
	// 检查迁移锁：迁移进行中禁止切换存储
	runningCount, err := l.migrationModel.GetRunningTaskCount(ctx)
	if err != nil {
		return fmt.Errorf("查询迁移任务状态失败: %w", err)
	}
	if runningCount > 0 {
		return fmt.Errorf("存储迁移进行中，禁止切换存储配置")
	}

	// 检查是否所有素材已迁移到目标平台
	notMigratedCount, err := l.mediaModel.CountNotOnStorage(ctx, provider)
	if err != nil {
		return fmt.Errorf("查询未迁移素材数量失败: %w", err)
	}
	if notMigratedCount > 0 {
		return fmt.Errorf("还有 %d 个素材未迁移到目标平台，请先在「素材迁移」中完成迁移后再激活", notMigratedCount)
	}

	if err := l.model.SetActive(ctx, provider); err != nil {
		return fmt.Errorf("激活存储配置失败: %w", err)
	}

	// 删除所有非激活的旧配置，确保只剩一个配置
	if err := l.model.DeleteNonActive(ctx); err != nil {
		return fmt.Errorf("清理旧配置失败: %w", err)
	}

	// 热重载存储管理器
	if err := l.manager.Reload(ctx); err != nil {
		return fmt.Errorf("存储配置已激活，但热重载失败: %w", err)
	}
	return nil
}

// TestConfig 测试存储连通性（不落库），AccessSecret 为明文。
// local 存储由 Manager 检查目录可写性。
func (l *StorageLogic) TestConfig(ctx context.Context, r *req.StorageTestReq) error {
	// 云存储校验必填字段
	if r.Provider != "local" {
		if r.Endpoint == "" {
			return fmt.Errorf("endpoint 不能为空")
		}
		if r.AccessKey == "" {
			return fmt.Errorf("access_key 不能为空")
		}
		if r.AccessSecret == "" {
			return fmt.Errorf("access_secret 不能为空")
		}
	}

	config := &model.StorageConfig{
		Provider:     r.Provider,
		Endpoint:     r.Endpoint,
		Region:       r.Region,
		Bucket:       r.Bucket,
		AccessKey:    r.AccessKey,
		AccessSecret: r.AccessSecret, // 明文，不落库
		PathPrefix:   r.PathPrefix,
		CustomDomain: r.CustomDomain,
		Extra:        r.Extra,
	}

	if err := l.manager.TestProvider(ctx, config); err != nil {
		return fmt.Errorf("存储连通性测试失败: %w", err)
	}
	return nil
}
