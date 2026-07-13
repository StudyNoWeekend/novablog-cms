package storage

import (
	"encoding/json"
	"fmt"

	"novablog/internal/model"
	"novablog/utils/crypto"
)

// NewProvider 根据 StorageConfig 构建对应 Provider。
// config.AccessSecret 是加密密文，需用 cryptoKey 解密。
// config.Extra 是 JSON，需解析出平台特有参数（如 MinIO 的 use_ssl）。
func NewProvider(config *model.StorageConfig, cryptoKey string) (StorageProvider, error) {
	plainSecret, err := crypto.Decrypt(config.AccessSecret, cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("解密 access_secret 失败: %w", err)
	}
	return buildProvider(config, plainSecret)
}

// buildProvider 根据配置和明文密钥构建 Provider。
// 该函数接收明文密钥，NewProvider 内部解密后调用本函数，
// TestProvider 直接传入明文密钥调用本函数，跳过解密步骤。
func buildProvider(config *model.StorageConfig, plainSecret string) (StorageProvider, error) {
	switch config.Provider {
	case "aliyun":
		return NewAliyunProvider(
			config.Endpoint, config.Region, config.Bucket,
			config.AccessKey, plainSecret,
			config.PathPrefix, config.CustomDomain,
		)
	case "tencent":
		return NewTencentProvider(
			config.Endpoint, config.AccessKey, plainSecret,
			config.Region, config.PathPrefix, config.CustomDomain,
		)
	case "minio":
		useSSL := parseUseSSL(config.Extra)
		return NewMinioProvider(
			config.Endpoint, config.Region, config.Bucket,
			config.AccessKey, plainSecret,
			config.PathPrefix, config.CustomDomain, useSSL,
		)
	default:
		return nil, fmt.Errorf("不支持的存储提供商: %s", config.Provider)
	}
}

// parseUseSSL 从 extra JSON 中解析 use_ssl 字段，默认 false。
func parseUseSSL(extra string) bool {
	if extra == "" {
		return false
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(extra), &m); err != nil {
		return false
	}
	if v, ok := m["use_ssl"]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
