package bootstrap

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"novablog/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 媒体引用启动回填集成测试使用本地 PostgreSQL（不可达时自动跳过）。
// 数据库独立命名（novablog_media_backfill_test），不会触碰开发/生产数据；
// 与 migrations/043 的手工执行逻辑一致，验证幂等与匹配规则。
const mediaBackfillTestDBName = "novablog_media_backfill_test"

var (
	mediaBackfillTestOnce  sync.Once
	mediaBackfillTestReady bool
	mediaBackfillTestDB    *gorm.DB
)

func setupMediaBackfillTest(t *testing.T) *gorm.DB {
	t.Helper()
	mediaBackfillTestOnce.Do(func() {
		user := os.Getenv("USER")
		baseDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=postgres sslmode=disable"
		base, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		base.Exec("DROP DATABASE IF EXISTS " + mediaBackfillTestDBName)
		if err := base.Exec("CREATE DATABASE " + mediaBackfillTestDBName).Error; err != nil {
			return
		}
		testDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=" + mediaBackfillTestDBName + " sslmode=disable"
		db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		if err := db.AutoMigrate(
			&model.Media{}, &model.MediaPreset{},
			&model.Article{}, &model.Song{}, &model.ThirdPartyPlaylist{}, &model.VideoWork{},
			&model.TravelGuide{}, &model.Recipe{}, &model.Book{}, &model.Game{},
			&model.PhotoEquipment{}, &model.Project{}, &model.TechStackItem{}, &model.Blogger{},
		); err != nil {
			return
		}
		mediaBackfillTestDB = db
		mediaBackfillTestReady = true
	})
	if !mediaBackfillTestReady {
		t.Skip("跳过媒体回填集成测试：本地 PostgreSQL 不可用")
	}
	return mediaBackfillTestDB
}

// TestMediaReferenceBackfill 验证启动回填：media.url 归一 + 业务表媒体引用匹配 + 外链保持 + 幂等。
func TestMediaReferenceBackfill(t *testing.T) {
	db := setupMediaBackfillTest(t)
	ctx := context.Background()
	logger := zap.NewNop()

	mediaID := uuid.NewString()
	presetID := uuid.NewString()
	mediaKey := "音乐/2026/01/" + mediaID + ".jpg"
	presetKey := "presets/" + mediaID + "/" + presetID + ".jpg"
	legacyMediaURL := "https://old-bucket.cos.example.com/" + presetKey // 注意：完整 URL 以 key 结尾
	legacyMediaURL = "https://old-bucket.cos.example.com/" + mediaKey
	legacyPresetURL := "https://old-bucket.cos.example.com/" + presetKey

	require.NoError(t, db.Exec(`DELETE FROM media WHERE id = ?`, mediaID).Error)
	require.NoError(t, db.Exec(`DELETE FROM media_presets WHERE id = ?`, presetID).Error)
	require.NoError(t, db.Exec(`DELETE FROM articles WHERE slug LIKE 'backfill-test-%'`).Error)
	require.NoError(t, db.Exec(`DELETE FROM bloggers WHERE username = 'backfill-test'`).Error)

	require.NoError(t, db.Create(&model.Media{
		ID: mediaID, Filename: "cover.jpg", URL: legacyMediaURL, StoragePath: mediaKey, StorageType: "tencent",
	}).Error)
	require.NoError(t, db.Create(&model.MediaPreset{
		ID: presetID, MediaID: mediaID, Name: "v1", FrameConfig: "{}", DisplayParams: "{}",
		OutputURL: legacyPresetURL, OutputStoragePath: presetKey,
	}).Error)

	externalAvatar := "https://simpleicons.org/icons/go.svg"
	blogger := model.Blogger{ID: uuid.NewString(), Username: "backfill-test", PasswordHash: "x", Avatar: externalAvatar}
	require.NoError(t, db.Create(&blogger).Error)

	articleByMedia := model.Article{ID: uuid.NewString(), Title: "按媒体回填", Slug: "backfill-test-a1", Content: "x", CoverImage: legacyMediaURL}
	articleByPreset := model.Article{ID: uuid.NewString(), Title: "按预设回填", Slug: "backfill-test-a2", Content: "x", CoverImage: legacyPresetURL}
	require.NoError(t, db.Create(&articleByMedia).Error)
	require.NoError(t, db.Create(&articleByPreset).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM media WHERE id = ?`, mediaID)
		db.Exec(`DELETE FROM media_presets WHERE id = ?`, presetID)
		db.Exec(`DELETE FROM articles WHERE slug LIKE 'backfill-test-%'`)
		db.Exec(`DELETE FROM bloggers WHERE username = 'backfill-test'`)
	})

	require.True(t, mediaBackfillNeeded(db), "初始应探测到待回填数据")

	normalizeMediaURLs(db, logger)
	backfillMediaRefFields(db, logger)

	// media.url / output_url 归一为相对 key
	var m model.Media
	require.NoError(t, db.Where("id = ?", mediaID).First(&m).Error)
	assert.Equal(t, mediaKey, m.URL, "media.url 应归一为相对 key")
	var p model.MediaPreset
	require.NoError(t, db.Where("id = ?", presetID).First(&p).Error)
	assert.Equal(t, "tencent", p.StorageType, "预设 storage_type 应继承所属媒体")
	assert.Equal(t, presetKey, p.OutputURL, "output_url 应归一为相对 key")

	// 业务表引用回填
	var a1 model.Article
	require.NoError(t, db.Where("id = ?", articleByMedia.ID).First(&a1).Error)
	require.NotNil(t, a1.CoverMediaID)
	assert.Equal(t, mediaID, *a1.CoverMediaID, "按完整 URL 匹配应回填 cover_media_id")
	var a2 model.Article
	require.NoError(t, db.Where("id = ?", articleByPreset.ID).First(&a2).Error)
	require.NotNil(t, a2.CoverPresetID)
	assert.Equal(t, presetID, *a2.CoverPresetID, "按预设成品图匹配应回填 cover_preset_id")

	// 外链保持 URL 模式
	var b model.Blogger
	require.NoError(t, db.Where("username = ?", "backfill-test").First(&b).Error)
	assert.Equal(t, externalAvatar, b.Avatar, "外链头像不应被改写")
	assert.Nil(t, b.AvatarMediaID, "外链头像不应回填媒体引用")

	// 幂等：重复执行不改变任何值
	normalizeMediaURLs(db, logger)
	backfillMediaRefFields(db, logger)
	var m2 model.Media
	require.NoError(t, db.Where("id = ?", mediaID).First(&m2).Error)
	assert.Equal(t, mediaKey, m2.URL, "重复归一应无变化")
	var a1b model.Article
	require.NoError(t, db.Where("id = ?", articleByMedia.ID).First(&a1b).Error)
	assert.Equal(t, mediaID, fmt.Sprintf("%v", deref(a1b.CoverMediaID)), "重复回填应无变化")

	// 上下文引用不产生副作用
	_ = ctx
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
