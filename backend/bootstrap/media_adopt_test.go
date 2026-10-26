package bootstrap

import (
	"os"
	"path/filepath"
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

// 遗留文件收编集成测试使用本地 PostgreSQL（不可达时自动跳过）。
const mediaAdoptTestDBName = "novablog_media_adopt_test"

var (
	mediaAdoptTestOnce  sync.Once
	mediaAdoptTestReady bool
	mediaAdoptTestDB    *gorm.DB
)

func setupMediaAdoptTest(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	mediaAdoptTestOnce.Do(func() {
		user := os.Getenv("USER")
		base, err := gorm.Open(postgres.Open("host=127.0.0.1 port=5432 user="+user+" dbname=postgres sslmode=disable"),
			&gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		base.Exec("DROP DATABASE IF EXISTS " + mediaAdoptTestDBName)
		if err := base.Exec("CREATE DATABASE " + mediaAdoptTestDBName).Error; err != nil {
			return
		}
		db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=5432 user="+user+" dbname="+mediaAdoptTestDBName+" sslmode=disable"),
			&gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
		if err != nil {
			return
		}
		// 迁移探测涉及的全部表(mediaBackfillNeeded 的 HasTable 前置检查要求齐备)
		if err := db.AutoMigrate(
			&model.Media{}, &model.MediaPreset{}, &model.MediaFolder{},
			&model.Article{}, &model.Song{}, &model.ThirdPartyPlaylist{}, &model.VideoWork{},
			&model.TravelGuide{}, &model.Recipe{}, &model.Book{}, &model.Game{},
			&model.PhotoEquipment{}, &model.Project{}, &model.TechStackItem{}, &model.Blogger{},
		); err != nil {
			return
		}
		mediaAdoptTestDB = db
		mediaAdoptTestReady = true
	})
	if !mediaAdoptTestReady {
		t.Skip("跳过遗留文件收编集成测试：本地 PostgreSQL 不可用")
	}
	uploadDir := t.TempDir()
	t.Cleanup(func() {
		mediaAdoptTestDB.Exec("DELETE FROM media")
		mediaAdoptTestDB.Exec("DELETE FROM media_folders")
		mediaAdoptTestDB.Exec("DELETE FROM bloggers")
	})
	return mediaAdoptTestDB, uploadDir
}

func adoptTestEnv(t *testing.T, db *gorm.DB) (bloggerID, profileFolderID string) {
	t.Helper()
	oldDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = oldDB })

	// 播种 profile 模块文件夹(验证收编归类)
	profileFolderID = uuid.NewString()
	require.NoError(t, db.Exec(
		`INSERT INTO media_folders (id, name, module_key, created_at) VALUES (?, '博主资料', 'profile', now())`,
		profileFolderID).Error)

	bloggerID = uuid.NewString()
	require.NoError(t, db.Create(&model.Blogger{
		ID:           bloggerID,
		Username:     "adopt-test",
		PasswordHash: "x",
		Avatar:       "http://old.example.com/files/profile/avatar/2026/09/aaa.jpeg",
		BlogIcon:     "http://old.example.com/files/profile/icon/2026/09/bbb.svg",
		// page_background 留空:验证空值行被天然跳过
	}).Error)
	return bloggerID, profileFolderID
}

// TestAdoptLegacyLocalFiles 验证遗留直写文件收编:
// 磁盘存在 → 补建 media 并归类文件夹、回填引用列;磁盘缺失 → 保持原样;幂等。
func TestAdoptLegacyLocalFiles(t *testing.T) {
	db, uploadDir := setupMediaAdoptTest(t)
	bloggerID, profileFolderID := adoptTestEnv(t, db)
	logger := zap.NewNop()

	// 磁盘:avatar 文件存在(icon 文件故意不放,验证磁盘缺失保持原样)
	avatarKey := "profile/avatar/2026/09/aaa.jpeg"
	avatarPath := filepath.Join(uploadDir, filepath.FromSlash(avatarKey))
	require.NoError(t, os.MkdirAll(filepath.Dir(avatarPath), 0o755))
	require.NoError(t, os.WriteFile(avatarPath, []byte("jpeg-bytes"), 0o644))

	adoptLegacyLocalFiles(db, logger, uploadDir)

	// avatar:收编成功
	var blogger model.Blogger
	require.NoError(t, db.Where("id = ?", bloggerID).First(&blogger).Error)
	require.NotNil(t, blogger.AvatarMediaID, "磁盘存在的头像应回填 media 引用")

	var media model.Media
	require.NoError(t, db.Where("id = ?", *blogger.AvatarMediaID).First(&media).Error)
	assert.Equal(t, avatarKey, media.StoragePath)
	assert.Equal(t, avatarKey, media.URL, "url 应存相对 key")
	assert.Equal(t, "local", media.StorageType)
	assert.Equal(t, "aaa.jpeg", media.Filename)
	assert.Equal(t, int16(1), media.FileType)
	require.NotNil(t, media.FolderID)
	assert.Equal(t, profileFolderID, *media.FolderID, "应收编到博主资料文件夹")

	// icon:磁盘缺失,保持 URL 模式
	assert.Nil(t, blogger.BlogIconMediaID, "磁盘缺失的文件不应收编")
	assert.Equal(t, "http://old.example.com/files/profile/icon/2026/09/bbb.svg", blogger.BlogIcon)

	// 幂等:重复执行不新增 media、不改变引用
	adoptLegacyLocalFiles(db, logger, uploadDir)
	var mediaCount int64
	db.Model(&model.Media{}).Count(&mediaCount)
	assert.EqualValues(t, 1, mediaCount, "重复收编不应新建 media 记录")
	var blogger2 model.Blogger
	require.NoError(t, db.Where("id = ?", bloggerID).First(&blogger2).Error)
	assert.Equal(t, *blogger.AvatarMediaID, *blogger2.AvatarMediaID)
}

// TestAdoptLinkExistingMedia 验证 media 已有同 storage_path 记录时仅回填引用列。
func TestAdoptLinkExistingMedia(t *testing.T) {
	db, uploadDir := setupMediaAdoptTest(t)
	bloggerID, _ := adoptTestEnv(t, db)
	logger := zap.NewNop()

	existingID := uuid.NewString()
	require.NoError(t, db.Create(&model.Media{
		ID:          existingID,
		Filename:    "aaa.jpeg",
		FileType:    1,
		URL:         "profile/avatar/2026/09/aaa.jpeg",
		StoragePath: "profile/avatar/2026/09/aaa.jpeg",
		StorageType: "local",
	}).Error)

	adoptLegacyLocalFiles(db, logger, uploadDir)

	var blogger model.Blogger
	require.NoError(t, db.Where("id = ?", bloggerID).First(&blogger).Error)
	require.NotNil(t, blogger.AvatarMediaID)
	assert.Equal(t, existingID, *blogger.AvatarMediaID, "应关联到已存在的 media 记录而非新建")
	var count int64
	db.Model(&model.Media{}).Count(&count)
	assert.EqualValues(t, 1, count)
}

// TestExtractFilesKey 验证 key 提取:中文解码、查询串剥离、异常形态拒绝。
func TestExtractFilesKey(t *testing.T) {
	assert.Equal(t, "profile/avatar/2026/09/a.jpeg",
		extractFilesKey("http://old.example.com/files/profile/avatar/2026/09/a.jpeg"))
	assert.Equal(t, "未分类/2026/10/b.jpg",
		extractFilesKey("https://czfapi.novablog.cn/files/%E6%9C%AA%E5%88%86%E7%B1%BB/2026/10/b.jpg?v=1"))
	assert.Equal(t, "", extractFilesKey("https://cdn.example.com/img/a.png"), "非 /files/ 直写形态应拒绝")
	assert.Equal(t, "", extractFilesKey("http://evil.com/files/../../etc/passwd"), "路径穿越应拒绝")
	assert.Equal(t, "", extractFilesKey(""))
}

// TestMediaBackfillNeededWithFilter 回归测试:带软删除过滤的业务表存在待回填行时,
// 探测必须返回 true(修复前裸 SELECT 拼入 "t." 别名导致 SQL 错误、误判为无需回填)。
func TestMediaBackfillNeededWithFilter(t *testing.T) {
	db, _ := setupMediaAdoptTest(t)
	oldDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = oldDB })

	bloggerID := uuid.NewString()
	require.NoError(t, db.Create(&model.Blogger{
		ID:           bloggerID,
		Username:     "probe-test",
		PasswordHash: "x",
		Avatar:       "http://old.example.com/files/profile/avatar/2026/09/aaa.jpeg",
	}).Error)
	t.Cleanup(func() { db.Where("id = ?", bloggerID).Delete(&model.Blogger{}) })

	assert.True(t, mediaBackfillNeeded(db), "存在待回填行时探测必须为 true")
}
