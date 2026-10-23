package logic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
	"novablog/internal/storage"
	"novablog/pkg/clouddrive"
	"novablog/utils/crypto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/lib/pq" // 注册 database/sql 的 postgres 驱动
)

// 云盘导入集成测试使用本地 PostgreSQL（不可达时自动跳过）。
// 数据库独立命名（novablog_cloud_import_test），不会触碰开发/生产数据。
const cloudImportTestDBName = "novablog_cloud_import_test"

var (
	cloudImportTestOnce  sync.Once
	cloudImportTestReady bool
)

// setupCloudImportTest 初始化本地测试数据库与依赖（媒体表/存储降级/导入逻辑）。
func setupCloudImportTest(t *testing.T) (*CloudImportLogic, string) {
	t.Helper()
	cloudImportTestOnce.Do(func() {
		db, err := openCloudImportTestDB()
		if err != nil {
			t.Logf("跳过云盘导入集成测试：本地 PostgreSQL 不可用 (%v)", err)
			return
		}
		if err := db.AutoMigrate(
			&model.Media{}, &model.MediaFolder{},
			&model.CloudDriveAccount{}, &model.CloudImportTask{}, &model.CloudImportItem{},
		); err != nil {
			t.Logf("跳过云盘导入集成测试：迁移失败 (%v)", err)
			return
		}
		db.Exec("DELETE FROM media")
		db.Exec("DELETE FROM media_folders")
		db.Exec("DELETE FROM cloud_import_items")
		db.Exec("DELETE FROM cloud_import_tasks")
		db.Exec("DELETE FROM cloud_drive_accounts")
		model.DB = db
		cloudImportTestReady = true
	})
	if !cloudImportTestReady {
		t.Skip("本地 PostgreSQL 不可用，跳过集成测试")
	}

	// 每个用例从干净状态开始（测试共用同一测试库）
	cleanupCloudImportTables()

	uploadDir := t.TempDir()
	mediaLogic := NewMediaLogic(storage.NewManager(model.NewStorageConfig(), "test-key", uploadDir, zap.NewNop()), uploadDir)
	driveLogic := NewCloudDriveLogic("test-key", zap.NewNop())
	importLogic := NewCloudImportLogic(driveLogic, mediaLogic, zap.NewNop())
	return importLogic, uploadDir
}

// cleanupCloudImportTables 清空云盘导入相关测试表。
func cleanupCloudImportTables() {
	model.DB.Exec("DELETE FROM media")
	model.DB.Exec("DELETE FROM media_folders")
	model.DB.Exec("DELETE FROM cloud_import_items")
	model.DB.Exec("DELETE FROM cloud_import_tasks")
	model.DB.Exec("DELETE FROM cloud_drive_accounts")
}

// openCloudImportTestDB 连接（必要时创建）本地测试数据库。
func openCloudImportTestDB() (*gorm.DB, error) {
	user := os.Getenv("USER")
	dsn := func(dbname string) string {
		return fmt.Sprintf("host=127.0.0.1 port=5432 user=%s dbname=%s sslmode=disable", user, dbname)
	}

	// 确保测试库存在
	admin, err := gorm.Open(postgres.Open(dsn("postgres")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	var exists bool
	admin.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", cloudImportTestDBName).Scan(&exists)
	if !exists {
		if err := admin.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cloudImportTestDBName)).Error; err != nil {
			return nil, err
		}
	}
	if sqlDB, err := admin.DB(); err == nil {
		sqlDB.Close()
	}

	return gorm.Open(postgres.Open(dsn(cloudImportTestDBName)), &gorm.Config{
		Logger: logger.Discard,
	})
}

// newTestWebDAVAccount 创建指向本地 httptest 服务器的 WebDAV 测试账号。
func newTestWebDAVAccount(t *testing.T, baseURL string) *model.CloudDriveAccount {
	t.Helper()
	encrypted, err := crypto.Encrypt("pass", "test-key")
	require.NoError(t, err)
	account := &model.CloudDriveAccount{
		ID:          uuid.New().String(),
		Provider:    clouddrive.ProviderWebDAV,
		OpenUserID:  webDAVOpenUserID(baseURL, "user"),
		DisplayName: "user",
		AccessToken: encrypted,
		BaseURL:     baseURL,
		Username:    "user",
		Status:      model.CloudAccountStatusActive,
	}
	require.NoError(t, model.NewCloudDriveAccount().Create(context.Background(), account))
	return account
}

// newTestID 生成测试用 UUID。
func newTestID() string {
	return uuid.New().String()
}

// TestStartImport_BatchLimit 验证单次导入数量上限校验。
func TestStartImport_BatchLimit(t *testing.T) {
	importLogic, _ := setupCloudImportTest(t)

	prev := *cloudDriveSettings
	defer SetCloudDriveSettings(&prev)
	SetCloudDriveSettings(&CloudDriveSettings{ImportBatchLimit: 2, MaxFileSizeMB: 2048})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	account := newTestWebDAVAccount(t, srv.URL)

	files := make([]req.CloudFile, 3)
	for i := range files {
		files[i] = req.CloudFile{FileID: fmt.Sprintf("/f%d.jpg", i), Name: fmt.Sprintf("f%d.jpg", i), Size: 10}
	}
	_, err := importLogic.StartImport(context.Background(), &req.CloudImportStartReq{
		AccountID: account.ID,
		Files:     files,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "超过上限")
}

// TestStartImport_FileTooLarge 验证单文件大小上限校验。
func TestStartImport_FileTooLarge(t *testing.T) {
	importLogic, _ := setupCloudImportTest(t)

	prev := *cloudDriveSettings
	defer SetCloudDriveSettings(&prev)
	SetCloudDriveSettings(&CloudDriveSettings{ImportBatchLimit: 200, MaxFileSizeMB: 1})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	account := newTestWebDAVAccount(t, srv.URL)

	_, err := importLogic.StartImport(context.Background(), &req.CloudImportStartReq{
		AccountID: account.ID,
		Files:     []req.CloudFile{{FileID: "/big.mp4", Name: "big.mp4", Size: 10 << 20}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "大小上限")
}

// TestStartImport_RunningTaskLock 验证同账号并发任务锁。
func TestStartImport_RunningTaskLock(t *testing.T) {
	importLogic, _ := setupCloudImportTest(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	account := newTestWebDAVAccount(t, srv.URL)

	// 预置一个 running 任务
	require.NoError(t, model.NewCloudImport().CreateTask(ctx, &model.CloudImportTask{
		ID:        newTestID(),
		AccountID: account.ID,
		Provider:  clouddrive.ProviderWebDAV,
		Status:    "running",
		Total:     1,
	}))

	_, err := importLogic.StartImport(ctx, &req.CloudImportStartReq{
		AccountID: account.ID,
		Files:     []req.CloudFile{{FileID: "/a.jpg", Name: "a.jpg", Size: 10}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "已有导入任务")
}

// TestStartImport_EndToEnd 完整链路：httptest WebDAV → 下载 → 本地存储入库 → 计数器回填，
// 并验证下载中的逐文件进度（transferred_bytes）可见。
func TestStartImport_EndToEnd(t *testing.T) {
	importLogic, uploadDir := setupCloudImportTest(t)
	ctx := context.Background()

	// a.jpg 分两段返回：先给前半，等测试观察到中间进度后再放行后半
	body := "jpeg-data-alpha"
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "user" || password != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "a.jpg") {
			_, _ = w.Write([]byte("png-data-beta"))
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = w.Write([]byte(body[:8]))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		<-release
		_, _ = w.Write([]byte(body[len(body)-7:]))
	}))
	defer srv.Close()
	account := newTestWebDAVAccount(t, srv.URL)

	// 调小进度回写阈值：httptest 文件仅 15 字节，默认 1MiB 阈值下中间块不会触发回写
	prevFlushBytes, prevFlushInterval := progressFlushBytes, progressFlushInterval
	progressFlushBytes, progressFlushInterval = 1, 0
	t.Cleanup(func() {
		progressFlushBytes, progressFlushInterval = prevFlushBytes, prevFlushInterval
	})

	taskID, err := importLogic.StartImport(ctx, &req.CloudImportStartReq{
		AccountID: account.ID,
		Files: []req.CloudFile{
			{FileID: "/photos/a.jpg", Path: "/photos/a.jpg", Name: "a.jpg", Size: int64(len(body))},
			{FileID: "/photos/b.png", Path: "/photos/b.png", Name: "b.png", Size: int64(len("png-data-beta"))},
		},
	})
	require.NoError(t, err)

	// 阶段一：等第一个文件出现中间进度（transferred_bytes > 0 且未完成）
	var midProgress bool
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		result, err := importLogic.GetImportStatus(ctx, taskID)
		require.NoError(t, err)
		if len(result.Items) > 0 && result.Items[0].TransferredBytes > 0 {
			midProgress = true
			break
		}
		if result.Task.Status == "completed" || result.Task.Status == "failed" {
			break // 进度过快已跑完（不应发生：handler 卡在 release）
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.True(t, midProgress, "下载进行中应观察到条目进度（transferred_bytes > 0）")
	close(release)

	// 阶段二：等待任务完成
	var statusRes *res.CloudImportStatusRes
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		result, err := importLogic.GetImportStatus(ctx, taskID)
		require.NoError(t, err)
		if result.Task.Status == "completed" || result.Task.Status == "failed" {
			statusRes = result
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NotNil(t, statusRes, "导入任务未在时限内完成")
	require.Equal(t, "completed", statusRes.Task.Status, "失败明细: %+v", statusRes.Failed)
	assert.Equal(t, 2, statusRes.Task.Succeeded)
	assert.Equal(t, 0, statusRes.Task.Failed)
	assert.Empty(t, statusRes.Failed)

	// 完成后每个条目的 transferred_bytes 应等于文件大小
	require.Len(t, statusRes.Items, 2)
	for _, item := range statusRes.Items {
		assert.Equal(t, item.FileSize, item.TransferredBytes, "file: "+item.Filename)
		assert.NotEmpty(t, item.MediaID)
	}

	// 校验媒体入库：文件名保留、本地落盘、大小与内容正确
	mediaList, _, err := model.NewMedia().GetList(ctx, nil, nil, nil, false, 1, 10)
	require.NoError(t, err)
	require.Len(t, mediaList, 2)
	byName := map[string]model.Media{}
	for _, m := range mediaList {
		byName[m.Filename] = m
	}
	a := byName["a.jpg"]
	require.NotEmpty(t, a.ID)
	assert.Equal(t, int64(len(body)), a.Size)
	assert.Equal(t, "local", a.StorageType)
	content, err := os.ReadFile(filepath.Join(uploadDir, a.StoragePath))
	require.NoError(t, err)
	assert.Equal(t, body, string(content))

	// 条目回填 media_id
	items, err := model.NewCloudImport().GetItemsByTaskID(ctx, taskID)
	require.NoError(t, err)
	for _, item := range items {
		assert.NotEmpty(t, item.MediaID)
	}
}

// TestRecoverRunningTasks 验证启动自愈：running 任务被标记为失败。
func TestRecoverRunningTasks(t *testing.T) {
	importLogic, _ := setupCloudImportTest(t)
	ctx := context.Background()
	importModel := model.NewCloudImport()

	require.NoError(t, importModel.CreateTask(ctx, &model.CloudImportTask{
		ID: newTestID(), AccountID: newTestID(), Provider: "webdav", Status: "running", Total: 1,
	}))
	require.NoError(t, importModel.CreateTask(ctx, &model.CloudImportTask{
		ID: newTestID(), AccountID: newTestID(), Provider: "baidu", Status: "completed", Total: 1,
	}))

	importLogic.RecoverInterruptedTasks(ctx)

	tasks, _, err := importModel.GetTasks(ctx, 1, 50)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(tasks), 2)
	for _, task := range tasks {
		assert.NotEqual(t, "running", task.Status)
		if task.Status == "failed" {
			assert.Contains(t, task.Error, "服务重启")
		}
	}
}
