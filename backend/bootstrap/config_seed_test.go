package bootstrap

import (
	"os"
	"sync"
	"testing"

	"novablog/internal/model"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 运行时配置种子集成测试使用本地 PostgreSQL（不可达时自动跳过）。
// 数据库独立命名（novablog_config_seed_test），不会触碰开发/生产数据。
const configSeedTestDBName = "novablog_config_seed_test"

var (
	configSeedTestOnce  sync.Once
	configSeedTestReady bool
	configSeedTestDB    *gorm.DB
)

func setupConfigSeedTest(t *testing.T) *gorm.DB {
	t.Helper()
	configSeedTestOnce.Do(func() {
		user := os.Getenv("USER")
		baseDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=postgres sslmode=disable"
		base, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		base.Exec("DROP DATABASE IF EXISTS " + configSeedTestDBName)
		if err := base.Exec("CREATE DATABASE " + configSeedTestDBName).Error; err != nil {
			return
		}
		testDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=" + configSeedTestDBName + " sslmode=disable"
		db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		if err := db.AutoMigrate(&model.SystemConfig{}); err != nil {
			return
		}
		configSeedTestDB = db
		configSeedTestReady = true
	})
	if !configSeedTestReady {
		t.Skip("跳过配置种子集成测试：本地 PostgreSQL 不可用")
	}
	return configSeedTestDB
}

func seedTestViper(t *testing.T) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.Set("upload.base_url", "https://blog.example.com")
	v.Set("themes.proxy_url", "http://proxy.example.com:7890")
	v.Set("image_search.proxy_url", "")
	v.Set("themes.github_token", "ghp_test_token")
	v.Set("themes.max_artifact_mb", 150)
	v.Set("clouddrive.max_file_size_mb", 1024)
	v.Set("clouddrive.import_batch_limit", 50)
	v.Set("geoip.download_url", "")
	return v
}

func seedConfigValue(t *testing.T, db *gorm.DB, key string) string {
	t.Helper()
	var cfg model.SystemConfig
	err := db.Where("\"key\" = ?", key).First(&cfg).Error
	require.NoError(t, err, "配置键 %s 应已播种", key)
	return cfg.Value
}

// TestSeedSystemConfigs 验证存量 yaml 键幂等导入：多代理键归一、空值不播种、重复执行不覆盖。
func TestSeedSystemConfigs(t *testing.T) {
	db := setupConfigSeedTest(t)
	oldDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = oldDB })

	v := seedTestViper(t)
	logger := zap.NewNop()

	seedSystemConfigs(v, logger)

	// 多代理键归一：取第一个非空（themes.proxy_url）
	assert.Equal(t, "http://proxy.example.com:7890", seedConfigValue(t, db, model.SystemConfigKeyProxyURL))
	assert.Equal(t, "https://blog.example.com", seedConfigValue(t, db, model.SystemConfigKeySiteURL))
	assert.Equal(t, "ghp_test_token", seedConfigValue(t, db, model.SystemConfigKeyGithubToken))
	assert.Equal(t, "150", seedConfigValue(t, db, model.SystemConfigKeyMaxArtifactMB))
	assert.Equal(t, "1024", seedConfigValue(t, db, model.SystemConfigKeyCloudMaxFileSizeMB))
	assert.Equal(t, "50", seedConfigValue(t, db, model.SystemConfigKeyCloudImportBatchLimit))
	// 空值不播种
	var count int64
	db.Model(&model.SystemConfig{}).Where("\"key\" = ?", model.SystemConfigKeyGeoipDownloadURL).Count(&count)
	assert.Zero(t, count, "空值键不应播种")

	// 幂等：DB 已有值后，yaml 变化不再覆盖（数据库为唯一主源）
	v.Set("themes.proxy_url", "http://changed.example.com:9999")
	v.Set("upload.base_url", "https://changed.example.com")
	seedSystemConfigs(v, logger)
	assert.Equal(t, "http://proxy.example.com:7890", seedConfigValue(t, db, model.SystemConfigKeyProxyURL),
		"重复播种不应覆盖 DB 已有值")
	assert.Equal(t, "https://blog.example.com", seedConfigValue(t, db, model.SystemConfigKeySiteURL),
		"重复播种不应覆盖 DB 已有值")

	// 缺失键补齐：删除一个键后重播种应恢复（用于手工清理后自愈）
	db.Where("\"key\" = ?", model.SystemConfigKeyGithubToken).Delete(&model.SystemConfig{})
	seedSystemConfigs(v, logger)
	assert.Equal(t, "ghp_test_token", seedConfigValue(t, db, model.SystemConfigKeyGithubToken),
		"缺失键应重新播种")
}
