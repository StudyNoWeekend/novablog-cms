package logic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	"novablog/internal/dto/req"
	"novablog/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/lib/pq" // 注册 database/sql 的 postgres 驱动
)

// 集成测试使用本地 PostgreSQL（brew 默认安装即可），不可达时自动跳过。
// 数据库独立命名（novablog_media_folder_test），不会触碰开发/生产数据。
const mediaFolderTestDBName = "novablog_media_folder_test"

var (
	mediaFolderTestOnce  sync.Once
	mediaFolderTestReady bool
)

// setupMediaFolderTest 初始化本地测试数据库（仅迁移媒体相关表），不可用时跳过测试。
func setupMediaFolderTest(t *testing.T) context.Context {
	t.Helper()
	mediaFolderTestOnce.Do(func() {
		db, err := openMediaFolderTestDB()
		if err != nil {
			t.Logf("跳过媒体文件夹集成测试：本地 PostgreSQL 不可用 (%v)", err)
			return
		}
		if err := db.AutoMigrate(&model.MediaFolder{}, &model.Media{}); err != nil {
			t.Logf("跳过媒体文件夹集成测试：迁移失败 (%v)", err)
			return
		}
		// 专属测试库，直接清空保证用例从干净状态开始
		db.Exec("DELETE FROM media")
		db.Exec("DELETE FROM media_folders")
		model.DB = db
		mediaFolderTestReady = true
	})
	if !mediaFolderTestReady {
		t.Skip("本地 PostgreSQL 不可用，跳过集成测试")
	}
	return context.Background()
}

// openMediaFolderTestDB 连接（必要时创建）本地测试数据库。
func openMediaFolderTestDB() (*gorm.DB, error) {
	dsn := func(dbname string) string {
		return fmt.Sprintf("host=127.0.0.1 port=5432 user=%s dbname=%s sslmode=disable", os.Getenv("USER"), dbname)
	}

	db, err := gorm.Open(postgres.Open(dsn(mediaFolderTestDBName)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err == nil {
		return db, nil
	}

	// 数据库不存在则先连接默认库创建
	srv, err := sql.Open("postgres", dsn("postgres"))
	if err != nil {
		return nil, err
	}
	defer srv.Close()
	if _, err := srv.Exec(fmt.Sprintf("CREATE DATABASE %q", mediaFolderTestDBName)); err != nil {
		return nil, err
	}
	return gorm.Open(postgres.Open(dsn(mediaFolderTestDBName)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// TestMediaFolderSystemProtection 集成测试：系统文件夹（ModuleKey 非空）禁止重命名/移动/删除，
// 普通用户文件夹不受影响。
func TestMediaFolderSystemProtection(t *testing.T) {
	ctx := setupMediaFolderTest(t)
	folderModel := model.NewMediaFolder()

	// 模块 key 用随机 UUID，保证与真实模块零冲突
	moduleKey := uuid.NewString()
	sysFolder := &model.MediaFolder{
		ID:        uuid.NewString(),
		Name:      fmt.Sprintf("测试系统文件夹_%s", uuid.NewString()[:8]),
		ModuleKey: &moduleKey,
	}
	normalFolder := &model.MediaFolder{
		ID:   uuid.NewString(),
		Name: fmt.Sprintf("测试普通文件夹_%s", uuid.NewString()[:8]),
	}
	assert.NoError(t, folderModel.Create(ctx, sysFolder))
	assert.NoError(t, folderModel.Create(ctx, normalFolder))
	// 系统文件夹无法走 logic.Delete，清理统一走模型层
	t.Cleanup(func() {
		_ = folderModel.Delete(ctx, sysFolder.ID)
		_ = folderModel.Delete(ctx, normalFolder.ID)
	})

	newName := fmt.Sprintf("改名_%s", uuid.NewString()[:8])

	// 1. 系统文件夹重命名被拒
	_, err := NewMediaFolderLogic().Update(ctx, sysFolder.ID, &req.MediaFolderUpdateReq{Name: &newName})
	assert.ErrorContains(t, err, "不支持重命名")

	// 2. 系统文件夹移动被拒（移入普通文件夹下）
	_, err = NewMediaFolderLogic().Update(ctx, sysFolder.ID, &req.MediaFolderUpdateReq{ParentID: &normalFolder.ID})
	assert.ErrorContains(t, err, "不支持移动")

	// 3. 系统文件夹移动到根目录（与当前一致）属幂等请求，放行
	rootParent := ""
	_, err = NewMediaFolderLogic().Update(ctx, sysFolder.ID, &req.MediaFolderUpdateReq{ParentID: &rootParent})
	assert.NoError(t, err)

	// 4. 系统文件夹删除被拒，且名称未被篡改
	err = NewMediaFolderLogic().Delete(ctx, sysFolder.ID)
	assert.ErrorContains(t, err, "不支持删除")
	kept, err := folderModel.GetByID(ctx, sysFolder.ID)
	if assert.NoError(t, err) {
		assert.Equal(t, sysFolder.Name, kept.Name, "被拒的重命名不应生效")
		assert.NotNil(t, kept.ModuleKey)
	}

	// 5. 普通文件夹重命名正常
	_, err = NewMediaFolderLogic().Update(ctx, normalFolder.ID, &req.MediaFolderUpdateReq{Name: &newName})
	assert.NoError(t, err)

	// 6. 普通文件夹移动正常（移入系统文件夹下不受限——受限的只是系统文件夹自身）
	_, err = NewMediaFolderLogic().Update(ctx, normalFolder.ID, &req.MediaFolderUpdateReq{ParentID: &sysFolder.ID})
	assert.NoError(t, err)
	moved, err := folderModel.GetByID(ctx, normalFolder.ID)
	if assert.NoError(t, err) {
		if assert.NotNil(t, moved.ParentID) {
			assert.Equal(t, sysFolder.ID, *moved.ParentID)
		}
	}

	// 7. 普通文件夹删除正常
	err = NewMediaFolderLogic().Delete(ctx, normalFolder.ID)
	assert.NoError(t, err)
}
