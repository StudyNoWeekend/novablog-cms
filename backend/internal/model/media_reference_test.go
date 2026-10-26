package model

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 媒体引用扫描/清理/解析集成测试使用本地 PostgreSQL（不可达时自动跳过）。
// 数据库独立命名（novablog_media_ref_test），不会触碰开发/生产数据。
const mediaRefTestDBName = "novablog_media_ref_test"

var (
	mediaRefTestOnce  sync.Once
	mediaRefTestReady bool
	mediaRefTestDB    *gorm.DB
)

func setupMediaRefTest(t *testing.T) *gorm.DB {
	t.Helper()
	mediaRefTestOnce.Do(func() {
		// 与 cloud_import_logic_test 一致：本地 PG 信任认证，用户取当前系统用户
		user := os.Getenv("USER")
		baseDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=postgres sslmode=disable"
		base, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		base.Exec("DROP DATABASE IF EXISTS " + mediaRefTestDBName)
		if err := base.Exec("CREATE DATABASE " + mediaRefTestDBName).Error; err != nil {
			return
		}
		testDSN := "host=127.0.0.1 port=5432 user=" + user + " dbname=" + mediaRefTestDBName + " sslmode=disable"
		db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		if err := db.AutoMigrate(
			&Media{}, &MediaPreset{}, &MediaFolder{},
			&Article{}, &Song{}, &ThirdPartyPlaylist{}, &VideoWork{},
			&TravelGuide{}, &Recipe{}, &Book{}, &Game{},
			&PhotoEquipment{}, &Project{}, &TechStackItem{}, &Blogger{},
			&Portfolio{}, &PortfolioItem{},
		); err != nil {
			return
		}
		mediaRefTestDB = db
		mediaRefTestReady = true
	})
	if !mediaRefTestReady {
		t.Skip("跳过媒体引用集成测试：本地 PostgreSQL 不可用")
	}
	return mediaRefTestDB
}

// mediaRefFixture 测试数据：一张原图 + 一个预设 + 引用它的文章行。
type mediaRefFixture struct {
	media   Media
	preset  MediaPreset
	article Article // ID 引用行
	legacy  Article // 旧 URL 行
	stable  Article // 正文稳定链接行
}

func seedMediaRefFixture(t *testing.T, db *gorm.DB) *mediaRefFixture {
	t.Helper()

	// Usage/Resolver 模型基于全局 DB 构建，测试期间指向测试库
	oldDB := DB
	DB = db
	t.Cleanup(func() { DB = oldDB })

	media := Media{
		ID:          uuid.NewString(),
		Filename:    "cover.jpg",
		URL:         "cat/2026/01/uuid-cover.jpg",
		StoragePath: "cat/2026/01/uuid-cover.jpg",
		StorageType: "local",
	}
	preset := MediaPreset{
		ID:                uuid.NewString(),
		MediaID:           media.ID,
		Name:              "默认无 EXIF",
		FrameConfig:       "{}",
		DisplayParams:     "{}",
		OutputURL:         "presets/" + media.ID + "/uuid-out.jpg",
		OutputStoragePath: "presets/" + media.ID + "/uuid-out.jpg",
		StorageType:       "local",
	}
	require.NoError(t, db.Create(&media).Error)
	require.NoError(t, db.Create(&preset).Error)

	// 文章：ID 引用（预设成品图）；legacy 行用旧 URL 全量字符串
	article := Article{
		ID: uuid.NewString(), Title: "ID 引用文章", Slug: "media-ref-a1", Content: "hello",
		CoverMediaID: &media.ID, CoverPresetID: &preset.ID,
	}
	legacyURL := "http://old.example.com/files/cat/2026/01/uuid-cover.jpg"
	legacy := Article{
		ID: uuid.NewString(), Title: "旧URL文章", Slug: "media-ref-a2",
		Content:    fmt.Sprintf("![img](%s)", legacyURL),
		CoverImage: legacyURL,
	}
	stable := Article{
		ID: uuid.NewString(), Title: "稳定链接文章", Slug: "media-ref-a3",
		Content: "![img](http://api.example.com/api/v1/public/media/" + media.ID + "/file)",
	}
	require.NoError(t, db.Create(&article).Error)
	require.NoError(t, db.Create(&legacy).Error)
	require.NoError(t, db.Create(&stable).Error)

	t.Cleanup(func() {
		db.Exec("DELETE FROM articles WHERE slug LIKE 'media-ref-%'")
		db.Exec("DELETE FROM media WHERE id = ?", media.ID)
		db.Exec("DELETE FROM media_presets WHERE id = ?", preset.ID)
	})

	return &mediaRefFixture{media: media, preset: preset, article: article, legacy: legacy, stable: stable}
}

// TestFindUsagesIDPrecise 验证引用扫描：ID 精确命中 + 稳定链接/旧 URL LIKE 兜底 + 去重。
func TestFindUsagesIDPrecise(t *testing.T) {
	db := setupMediaRefTest(t)
	fx := seedMediaRefFixture(t, db)

	usage := NewMediaUsage()
	hits, err := usage.FindUsages(context.Background(),
		fx.media.ID, fx.media.URL, fx.media.StoragePath,
		[]string{fx.preset.ID}, []string{fx.preset.OutputStoragePath})
	require.NoError(t, err)

	byKey := make(map[string]bool, len(hits))
	for _, h := range hits {
		byKey[h.Module+"|"+h.ID+"|"+h.Field] = true
	}
	// ID 精确：文章封面（预设引用）
	assert.True(t, byKey["文章|"+fx.article.ID+"|封面"], "应精确命中 *_preset_id 引用: %v", hits)
	// LIKE 兜底：旧 URL 全等字段 + 正文旧 URL + 正文稳定链接
	assert.True(t, byKey["文章|"+fx.legacy.ID+"|封面"], "应命中旧 URL 封面: %v", hits)
	assert.True(t, byKey["文章|"+fx.legacy.ID+"|正文"], "应命中正文旧 URL: %v", hits)
	assert.True(t, byKey["文章|"+fx.stable.ID+"|正文"], "应命中正文稳定链接: %v", hits)
	// 去重：同一行不应出现重复命中
	assert.Len(t, hits, len(byKey), "命中应已去重: %v", hits)
}

// TestClearMediaReferences 验证强制删除前的引用列清理。
func TestClearMediaReferences(t *testing.T) {
	db := setupMediaRefTest(t)
	fx := seedMediaRefFixture(t, db)

	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return clearMediaReferences(tx, fx.media.ID, []string{fx.preset.ID})
	}))

	var updated Article
	require.NoError(t, db.Where("id = ?", fx.article.ID).First(&updated).Error)
	assert.Nil(t, updated.CoverMediaID, "media 引用应被清空")
	assert.Nil(t, updated.CoverPresetID, "preset 引用应被清空")
}

// TestResolveCoverURLsPriority 验证解析优先级：preset_id → media_id → 历史落库 URL。
func TestResolveCoverURLsPriority(t *testing.T) {
	db := setupMediaRefTest(t)
	fx := seedMediaRefFixture(t, db)
	ctx := context.Background()

	presetID := fx.preset.ID
	mediaID := fx.media.ID
	onlyMedia := uuid.NewString()
	articles := []*Article{
		{ID: fx.article.ID, CoverMediaID: &mediaID, CoverPresetID: &presetID},
		{ID: onlyMedia, CoverMediaID: &mediaID},
		{ID: uuid.NewString(), CoverImage: "http://external.example.com/a.png"},
	}

	ResolveCoverURLs(ctx, articles,
		func(a *Article) (string, string) { return DerefStr(a.CoverMediaID), DerefStr(a.CoverPresetID) },
		func(a *Article, u string) { a.CoverImage = u },
	)

	// local 存储 + 未注入 BaseURL/ResolveStorageURL 时，AfterFind 后 URL 为相对 key 本身
	assert.Equal(t, fx.preset.OutputStoragePath, articles[0].CoverImage, "preset 优先")
	assert.Equal(t, fx.media.StoragePath, articles[1].CoverImage, "无 preset 时按 media 解析")
	assert.Equal(t, "http://external.example.com/a.png", articles[2].CoverImage, "外链保持原值")
}
