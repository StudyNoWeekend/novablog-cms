package logic

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"novablog/internal/cache"
	"novablog/internal/dto/req"
	"novablog/internal/middleware"
	"novablog/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq" // 注册 database/sql 的 postgres 驱动
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 集成测试使用本地 PostgreSQL/Redis（brew 默认安装即可），不可达时自动跳过。
// 数据库独立命名（novablog_cors_test），不会触碰开发/生产数据。
const (
	corsTestDBName = "novablog_cors_test"
	corsTestRedisDB = 15
)

var (
	corsTestOnce  sync.Once
	corsTestDB    *gorm.DB
	corsTestRedis *redis.Client
	corsTestReady bool
)

// setupCorsTest 初始化测试环境（DB + Redis + 全局变量），不可用时跳过测试。
func setupCorsTest(t *testing.T) context.Context {
	t.Helper()
	corsTestOnce.Do(func() {
		db, err := openCorsTestDB()
		if err != nil {
			t.Logf("跳过跨域配置集成测试：本地 PostgreSQL 不可用 (%v)", err)
			return
		}
		rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: corsTestRedisDB})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			t.Logf("跳过跨域配置集成测试：本地 Redis 不可用 (%v)", err)
			return
		}
		if err := db.AutoMigrate(&model.CorsConfig{}); err != nil {
			t.Logf("跳过跨域配置集成测试：迁移失败 (%v)", err)
			return
		}
		corsTestDB = db
		corsTestRedis = rdb
		// 注入全局依赖（与 bootstrap.InitDB / InitRedis 等效）
		model.DB = db
		cache.RedisClient = rdb
		corsTestReady = true
	})
	if !corsTestReady {
		t.Skip("本地 PostgreSQL/Redis 不可用，跳过集成测试")
	}

	ctx := context.Background()
	cleanupCorsTestData(ctx)
	t.Cleanup(func() { cleanupCorsTestData(ctx) })
	return ctx
}

// openCorsTestDB 连接（必要时创建）本地测试数据库。
func openCorsTestDB() (*gorm.DB, error) {
	dsn := func(dbname string) string {
		return fmt.Sprintf("host=127.0.0.1 port=5432 user=%s dbname=%s sslmode=disable", os.Getenv("USER"), dbname)
	}

	db, err := gorm.Open(postgres.Open(dsn(corsTestDBName)), &gorm.Config{
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
	if _, err := srv.Exec(fmt.Sprintf("CREATE DATABASE %q", corsTestDBName)); err != nil {
		return nil, err
	}
	return gorm.Open(postgres.Open(dsn(corsTestDBName)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// cleanupCorsTestData 清空配置表与缓存键，保证用例间互不干扰。
func cleanupCorsTestData(ctx context.Context) {
	if corsTestDB != nil {
		corsTestDB.Exec("DELETE FROM cors_configs")
	}
	if corsTestRedis != nil {
		corsTestRedis.Del(ctx, cache.CacheKeyCorsConfig)
	}
}

// assertOriginAllowed 构建仅挂载 CORS 中间件的路由，断言某来源的预检放行与否。
func assertOriginAllowed(t *testing.T, origin string, allowed bool) bool {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORSMiddleware())
	r.OPTIONS("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	httpReq := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	httpReq.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	if allowed {
		return assert.Equal(t, http.StatusNoContent, w.Code, "来源 %s 应放行", origin) &&
			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"), "来源 %s 应回显 ACAO", origin)
	}
	return assert.Equal(t, http.StatusForbidden, w.Code, "来源 %s 应被拒绝", origin)
}

// TestCorsInitFromFallback_SeedWhenNoRow DB 无记录：以 config.yaml 值播种并生效。
func TestCorsInitFromFallback_SeedWhenNoRow(t *testing.T) {
	ctx := setupCorsTest(t)

	logic := NewCorsConfigLogic()
	err := logic.InitFromFallback(ctx, "https://cfg-seed.example.com")
	assert.NoError(t, err)

	// DB 应落库播种值
	row, err := model.NewCorsConfig().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "https://cfg-seed.example.com", row.AllowedOrigins)

	// 中间件应以播种值放行
	assert.True(t, assertOriginAllowed(t, "https://cfg-seed.example.com", true))
	assert.True(t, assertOriginAllowed(t, "https://other.example.com", false))
}

// TestCorsInitFromFallback_BackfillUntouchedRow DB 仅有迁移播种的初始行（从未修改）：以 config.yaml 回填。
func TestCorsInitFromFallback_BackfillUntouchedRow(t *testing.T) {
	ctx := setupCorsTest(t)

	// 模拟迁移播种行：created_at == updated_at
	_, err := model.NewCorsConfig().CreateConfig(ctx, "http://localhost:5173,http://localhost:5174")
	assert.NoError(t, err)

	logic := NewCorsConfigLogic()
	err = logic.InitFromFallback(ctx, "https://cfg-backfill.example.com")
	assert.NoError(t, err)

	// 初始行应被 config.yaml 值回填
	row, err := model.NewCorsConfig().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "https://cfg-backfill.example.com", row.AllowedOrigins)
	assert.True(t, assertOriginAllowed(t, "https://cfg-backfill.example.com", true))
}

// TestCorsInitFromFallback_DBWinsWhenCustomized DB 有用户自定义配置：无条件以 DB 为准，config.yaml 不覆盖。
func TestCorsInitFromFallback_DBWinsWhenCustomized(t *testing.T) {
	ctx := setupCorsTest(t)

	// 模拟用户已在后台保存过：记录被修改（updated_at 晚于 created_at）
	_, err := model.NewCorsConfig().CreateConfig(ctx, "https://db-authority.example.com")
	assert.NoError(t, err)
	assert.NoError(t, corsTestDB.Exec("UPDATE cors_configs SET updated_at = updated_at + interval '1 second'").Error)

	logic := NewCorsConfigLogic()
	err = logic.InitFromFallback(ctx, "https://cfg-ignored.example.com")
	assert.NoError(t, err)

	// DB 值原样保留，config.yaml 值被忽略
	row, err := model.NewCorsConfig().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "https://db-authority.example.com", row.AllowedOrigins)
	assert.True(t, assertOriginAllowed(t, "https://db-authority.example.com", true))
	assert.True(t, assertOriginAllowed(t, "https://cfg-ignored.example.com", false))
}

// TestCorsInitFromFallback_EmptyFallback DB 无记录且 config.yaml 为空：不落库，走默认值。
func TestCorsInitFromFallback_EmptyFallback(t *testing.T) {
	ctx := setupCorsTest(t)

	// 模拟真实启动态：进程刚起、中间件尚未注入任何来源
	middleware.SetAllowedOrigins("")

	logic := NewCorsConfigLogic()
	err := logic.InitFromFallback(ctx, "")
	assert.NoError(t, err)

	// 不落库
	_, err = model.NewCorsConfig().GetConfig(ctx)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// 中间件回退硬编码默认值
	assert.True(t, assertOriginAllowed(t, "http://localhost:5173", true))
}

// TestCorsUpdateConfig_HotSync 保存配置：DB 落库、缓存刷新、中间件即时热更新。
func TestCorsUpdateConfig_HotSync(t *testing.T) {
	ctx := setupCorsTest(t)
	logic := NewCorsConfigLogic()

	origins := "https://hot-a.example.com,https://hot-b.example.com"
	err := logic.UpdateConfig(ctx, &req.UpdateCorsConfigReq{AllowedOrigins: &origins})
	assert.NoError(t, err)

	// DB 落库
	row, err := model.NewCorsConfig().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, origins, row.AllowedOrigins)

	// 中间件即时生效（无需重启）
	assert.True(t, assertOriginAllowed(t, "https://hot-a.example.com", true))
	assert.True(t, assertOriginAllowed(t, "https://hot-b.example.com", true))
	assert.True(t, assertOriginAllowed(t, "https://hot-c.example.com", false))

	// 再次修改立即覆盖旧白名单
	origins2 := "https://hot-c.example.com"
	err = logic.UpdateConfig(ctx, &req.UpdateCorsConfigReq{AllowedOrigins: &origins2})
	assert.NoError(t, err)
	assert.True(t, assertOriginAllowed(t, "https://hot-c.example.com", true))
	assert.True(t, assertOriginAllowed(t, "https://hot-a.example.com", false))

	// GetConfig 读回
	res, err := logic.GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, origins2, res.AllowedOrigins)
}

// TestCorsUpdateConfig_CreatesWhenNoRow DB 无记录时保存：直接创建而非报错。
func TestCorsUpdateConfig_CreatesWhenNoRow(t *testing.T) {
	ctx := setupCorsTest(t)
	logic := NewCorsConfigLogic()

	origins := "https://first-save.example.com"
	err := logic.UpdateConfig(ctx, &req.UpdateCorsConfigReq{AllowedOrigins: &origins})
	assert.NoError(t, err)

	row, err := model.NewCorsConfig().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "https://first-save.example.com", row.AllowedOrigins)
}

// TestCorsGetConfig_EmptyTable DB 无记录时读取：返回空配置而非报错。
func TestCorsGetConfig_EmptyTable(t *testing.T) {
	ctx := setupCorsTest(t)

	res, err := NewCorsConfigLogic().GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", res.AllowedOrigins)
}
