package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"novablog/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const communityPushTestDBName = "novablog_push_cms_test"

// fakeCommunity 模拟 NovaBlog 社区推送 API 的最小实现。
type fakeCommunity struct {
	server *httptest.Server

	mu             sync.Mutex
	lastUpsert     map[string]any
	upsertCount    int
	deleteCalled   []string
	nextID         uint
	result         string // created | updated
	lastVerifyCode string
	verifyCount    int
	verifyFailWith *struct {
		code int
		body string
	}
	failWith *struct {
		code int
		body string
	}
}

func newFakeCommunity() *fakeCommunity {
	fake := &fakeCommunity{nextID: 100, result: "created"}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/push/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer nbk_test_key" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 401, "message": "invalid key"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "message": "success",
			"data": map[string]any{"username": "alice", "blog_url": "https://blog.example.com", "credential_name": "测试"},
		})
	})

	mux.HandleFunc("/api/v1/verify/blog", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		code, _ := body["verify_code"].(string)
		fake.mu.Lock()
		fake.lastVerifyCode = code
		fake.verifyCount++
		fail := fake.verifyFailWith
		fake.verifyFailWith = nil
		fake.mu.Unlock()
		if fail != nil {
			w.WriteHeader(fail.code)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": fail.code, "message": fail.body})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "message": "success",
			"data": map[string]any{"verified": true, "username": "alice"},
		})
	})

	mux.HandleFunc("/api/v1/push/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if fail := fake.takeFail(); fail != nil {
			w.WriteHeader(fail.code)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": fail.code, "message": fail.body})
			return
		}
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		fake.mu.Lock()
		fake.lastUpsert = payload
		fake.upsertCount++
		fake.nextID++
		articleID := fake.nextID
		result := fake.result
		if result == "created" {
			fake.result = "updated" // 模拟社区：同 external_id 二次推送返回 updated
		}
		fake.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "message": "success",
			"data": map[string]any{"article_id": articleID, "external_id": payload["external_id"], "result": result},
		})
	})

	mux.HandleFunc("/api/v1/push/articles/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/push/articles/"), "/")
		fake.mu.Lock()
		fake.deleteCalled = append(fake.deleteCalled, parts[0])
		fake.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "message": "success",
			"data": map[string]any{"external_id": parts[0], "result": "deleted"},
		})
	})

	fake.server = httptest.NewServer(mux)
	return fake
}

func (f *fakeCommunity) takeFail() *struct {
	code int
	body string
} {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failWith == nil {
		return nil
	}
	fail := f.failWith
	f.failWith = nil
	return fail
}

func (f *fakeCommunity) setFail(code int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failWith = &struct {
		code int
		body string
	}{code, body}
}

func (f *fakeCommunity) setVerifyFail(code int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.verifyFailWith = &struct {
		code int
		body string
	}{code, body}
}

func (f *fakeCommunity) getLastVerifyCode() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastVerifyCode
}

func (f *fakeCommunity) getVerifyCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.verifyCount
}

func (f *fakeCommunity) getLastUpsert() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make(map[string]any, len(f.lastUpsert))
	for k, v := range f.lastUpsert {
		cp[k] = v
	}
	return cp
}

// setupCommunityPushTest 初始化测试库（本地 PostgreSQL，不可用时跳过）、注入全局依赖与假社区。
func setupCommunityPushTest(t *testing.T) (context.Context, *fakeCommunity, *gorm.DB) {
	t.Helper()

	db, err := openCommunityPushTestDB()
	if err != nil {
		t.Skipf("本地 PostgreSQL 不可用，跳过集成测试: %v", err)
	}
	require.NoError(t, db.AutoMigrate(
		&model.Article{}, &model.Tag{}, &model.Category{},
		&model.TravelGuide{}, &model.Recipe{},
		&model.Portfolio{}, &model.PortfolioItem{}, &model.MediaPreset{},
		&model.OpenSourceWork{},
		&model.CommunityPushConfig{}, &model.CommunitySyncRecord{},
	))

	prevDB := model.DB
	model.DB = db
	prevBase := OfficialSiteBaseURL()
	prevSite := getCommunityPushSettings().SiteURL
	prevKey := getCommunityPushKey()

	fake := newFakeCommunity()
	SetOfficialSiteBaseURL(fake.server.URL)
	SetCommunitySiteURL("https://blog.example.com")
	SetCommunityPushKey("nbk_test_key")
	model.BaseURL = "https://blog.example.com"

	ctx := context.Background()
	cleanupCommunityPushData(ctx, db)
	t.Cleanup(func() {
		cleanupCommunityPushData(context.Background(), db)
		model.DB = prevDB
		SetOfficialSiteBaseURL(prevBase)
		SetCommunitySiteURL(prevSite)
		SetCommunityPushKey(prevKey)
		fake.server.Close()
	})
	return ctx, fake, db
}

func cleanupCommunityPushData(ctx context.Context, db *gorm.DB) {
	for _, table := range []string{
		"community_sync_records", "community_push_configs", "article_tags", "articles",
		"travel_guides", "recipes", "portfolio_items", "portfolios", "media_presets", "open_source_works",
	} {
		db.WithContext(ctx).Exec("DELETE FROM " + table)
	}
}

func openCommunityPushTestDB() (*gorm.DB, error) {
	dsn := func(dbname string) string {
		return fmt.Sprintf("host=127.0.0.1 port=5432 user=%s dbname=%s sslmode=disable", os.Getenv("USER"), dbname)
	}
	silent := &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)}
	db, err := gorm.Open(gormPostgres.Open(dsn(communityPushTestDBName)), silent)
	if err == nil {
		return db, nil
	}
	srv, srvErr := openRootPostgres(dsn("postgres"))
	if srvErr != nil {
		return nil, err
	}
	defer srv.Close()
	if _, err := srv.Exec(fmt.Sprintf(`CREATE DATABASE %q`, communityPushTestDBName)); err != nil {
		return nil, err
	}
	return gorm.Open(gormPostgres.Open(dsn(communityPushTestDBName)), silent)
}

// —— 构建器单测 ——

func seedPushArticle(t *testing.T, db *gorm.DB) *model.Article {
	t.Helper()
	article := &model.Article{
		ID:         uuid.NewString(),
		Title:      "Go 工程实践",
		Slug:       "go-practice",
		Summary:    "",
		Content:    "# 开始\n\n看图 ![封面](uploads/img/a.png) 与 [链接](/about)。\n\n```go\ngo build\n```",
		CoverImage: "uploads/img/cover.png",
		Status:     2,
		Type:       1,
	}
	now := time.Now()
	article.PublishedAt = &now
	require.NoError(t, db.Create(article).Error)

	tag := &model.Tag{ID: uuid.NewString(), Name: "Go"}
	require.NoError(t, db.Create(tag).Error)
	require.NoError(t, db.Exec("INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?)", article.ID, tag.ID).Error)
	return article
}

func TestBuildArticlePushPayload(t *testing.T) {
	ctx, _, db := setupCommunityPushTest(t)
	article := seedPushArticle(t, db)

	built, err := BuildCommunityPayload(ctx, TargetTypeArticle, article.ID)
	require.NoError(t, err)
	payload := built.Payload

	assert.Equal(t, "article", payload.Type)
	assert.Equal(t, article.ID, payload.ExternalID)
	assert.Equal(t, "published", payload.Status)
	assert.Equal(t, "https://blog.example.com/files/uploads/img/cover.png", payload.CoverImage)
	assert.Equal(t, "开始\n\n看图  与 链接。", payload.Summary) // 摘要缺省时剥离 markdown 语法自动生成
	assert.Contains(t, payload.ContentMarkdown, "https://blog.example.com/files/uploads/img/a.png")
	assert.NotContains(t, payload.ContentMarkdown, "](/uploads/")
	assert.Equal(t, "https://blog.example.com/articles/go-practice", payload.SourceURL)
	assert.Equal(t, []string{"Go"}, payload.Tags)
	assert.False(t, payload.PublishedAt == "")
}

func TestBuildRecipePushPayload(t *testing.T) {
	ctx, _, db := setupCommunityPushTest(t)
	recipe := &model.Recipe{
		ID:          uuid.NewString(),
		Title:       "豉汁蒸排骨",
		Cover:       "uploads/img/ribs.png",
		Summary:     "",
		Ingredients: []map[string]any{{"name": "排骨", "amount": "500g"}, {"name": "豆豉", "amount": "2 勺"}},
		Steps:       []map[string]any{{"text": "腌制半小时"}, {"text": "蒸 25 分钟", "image": "uploads/img/step2.png"}},
		Difficulty:  2,
		Minutes:     40,
		Servings:    3,
		Tags:        "粤菜,家常菜",
		Status:      1,
	}
	require.NoError(t, db.Create(recipe).Error)

	built, err := BuildCommunityPayload(ctx, TargetTypeRecipe, recipe.ID)
	require.NoError(t, err)
	payload := built.Payload

	assert.Equal(t, "recipe", payload.Type)
	assert.Equal(t, "https://blog.example.com/files/uploads/img/ribs.png", payload.CoverImage)
	assert.Contains(t, payload.ContentMarkdown, "## 食材清单")
	assert.Contains(t, payload.ContentMarkdown, "| 排骨 | 500g |")
	assert.Contains(t, payload.ContentMarkdown, "## 做法步骤")
	assert.Contains(t, payload.ContentMarkdown, "https://blog.example.com/files/uploads/img/step2.png")
	assert.Equal(t, []string{"粤菜", "家常菜"}, payload.Tags)
	assert.Equal(t, "published", payload.Status)
	// extra 与 CMS 字段同名直传
	assert.Equal(t, 40, payload.Extra["minutes"])
	assert.Len(t, payload.Extra["ingredients"], 2)
}

func TestBuildOpenSourcePushPayload(t *testing.T) {
	ctx, _, db := setupCommunityPushTest(t)
	work := &model.OpenSourceWork{
		ID:       uuid.NewString(),
		Name:     "novablog-editor",
		RepoURL:  "https://github.com/example/editor",
		Summary:  "Markdown 编辑器组件",
		Readme:   "# Editor\n\n一个编辑器。",
		Language: "TypeScript",
		Topics:   "react,editor",
		Stars:    42,
		Status:   1,
	}
	require.NoError(t, db.Create(work).Error)

	built, err := BuildCommunityPayload(ctx, TargetTypeOpenSource, work.ID)
	require.NoError(t, err)
	payload := built.Payload

	// slug 映射：open_source → opensource
	assert.Equal(t, "opensource", payload.Type)
	assert.Equal(t, work.RepoURL, payload.SourceURL) // 开源作品原文链接指向仓库
	assert.Equal(t, []string{"react", "editor"}, payload.Tags)
	assert.Equal(t, 42, payload.Extra["stars"])
	assert.Equal(t, "TypeScript", payload.Extra["language"])
	assert.Contains(t, payload.ContentMarkdown, "一个编辑器。")
}

func TestBuildTravelPushPayload(t *testing.T) {
	ctx, _, db := setupCommunityPushTest(t)
	guide := &model.TravelGuide{
		ID:          uuid.NewString(),
		Title:       "大理三日",
		Summary:     "洱海与古城",
		Destination: "大理",
		Region:      "云南",
		Days:        3,
		BestMonth:   "5-6 月",
		Status:      2,
		Attractions: []map[string]any{{"name": "才村码头", "description": "日落机位"}},
		Itinerary:   []map[string]any{{"day": float64(1), "title": "环湖", "description": "骑行"}},
	}
	require.NoError(t, db.Create(guide).Error)

	built, err := BuildCommunityPayload(ctx, TargetTypeTravel, guide.ID)
	require.NoError(t, err)
	payload := built.Payload

	assert.Equal(t, "travel", payload.Type)
	assert.Contains(t, payload.ContentMarkdown, "目的地 大理")
	assert.Contains(t, payload.ContentMarkdown, "## 景点推荐")
	assert.Contains(t, payload.ContentMarkdown, "### Day 1：环湖")
	assert.Equal(t, "大理", payload.Extra["destination"])
	assert.Len(t, payload.Extra["attractions"], 1)
	assert.Equal(t, "published", payload.Status)
}

func TestNormalizePushTagsTruncation(t *testing.T) {
	warnings := make([]string, 0)
	tags := make([]string, 0, 12)
	for i := 0; i < 12; i++ {
		tags = append(tags, fmt.Sprintf("标签%d", i))
	}
	result := normalizePushTags(tags, &warnings)
	assert.Len(t, result, 10)
	assert.NotEmpty(t, warnings)

	// 超长标签截断
	long := normalizePushTags([]string{strings.Repeat("长", 40)}, nil)
	assert.Equal(t, 30, utf8.RuneCountInString(long[0]))
}

// —— 推送服务状态机 ——

func TestCommunityPushLifecycle(t *testing.T) {
	ctx, fake, db := setupCommunityPushTest(t)
	article := seedPushArticle(t, db)
	l := NewCommunityPushLogic()

	t.Run("first push creates and records synced", func(t *testing.T) {
		result, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		assert.Equal(t, "created", result.Result)
		assert.Equal(t, model.CommunitySyncStatusSynced, result.SyncStatus)
		assert.Equal(t, uint(101), result.CommunityID)

		record, err := l.syncModel.GetByTarget(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, model.CommunitySyncStatusSynced, record.SyncStatus)
		assert.Equal(t, uint(101), record.CommunityArticleID)
	})

	t.Run("unchanged content skips upstream call", func(t *testing.T) {
		countBefore := fake.upsertCount
		result, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		assert.Equal(t, "skipped", result.Result)
		assert.Equal(t, countBefore, fake.upsertCount)
	})

	t.Run("content change pushes update", func(t *testing.T) {
		require.NoError(t, db.Model(&model.Article{}).Where("id = ?", article.ID).
			Update("title", "Go 工程实践（第二版）").Error)
		result, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated", result.Result)
		assert.Equal(t, "Go 工程实践（第二版）", fake.getLastUpsert()["title"])
	})

	t.Run("hide syncs hidden status", func(t *testing.T) {
		result, err := l.HideFromCommunity(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		assert.Equal(t, model.CommunitySyncStatusHidden, result.SyncStatus)
		last := fake.getLastUpsert()
		assert.Equal(t, "hidden", last["status"])
	})

	t.Run("push failure records failed for retry", func(t *testing.T) {
		fake.setFail(400, "extra 超过 64KB 上限")
		_, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
		require.Error(t, err)
		record, err2 := l.syncModel.GetByTarget(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err2)
		assert.Equal(t, model.CommunitySyncStatusFailed, record.SyncStatus)
		assert.Contains(t, record.LastError, "extra")
		assert.Equal(t, 1, record.RetryCount)
	})

	t.Run("retry worker pushes latest content", func(t *testing.T) {
		retried, err := l.RetryFailed(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, retried)
		record, err2 := l.syncModel.GetByTarget(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err2)
		assert.Equal(t, model.CommunitySyncStatusSynced, record.SyncStatus)
	})

	t.Run("delete syncs removal", func(t *testing.T) {
		result, err := l.RemoveFromCommunity(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err)
		assert.Equal(t, "deleted", result.Result)
		assert.Contains(t, fake.deleteCalled, article.ID)
		record, err2 := l.syncModel.GetByTarget(ctx, TargetTypeArticle, article.ID)
		require.NoError(t, err2)
		assert.Equal(t, model.CommunitySyncStatusDeleted, record.SyncStatus)
	})

	t.Run("sync on updated is noop after deletion", func(t *testing.T) {
		l.SyncOnUpdated(ctx, TargetTypeArticle, article.ID)
		assert.Equal(t, "deleted", func() string {
			record, _ := l.syncModel.GetByTarget(ctx, TargetTypeArticle, article.ID)
			return record.SyncStatus
		}())
	})
}

func TestCommunityPushNotConfigured(t *testing.T) {
	ctx, _, db := setupCommunityPushTest(t)
	article := seedPushArticle(t, db)

	prevBase := OfficialSiteBaseURL()
	prevKey := getCommunityPushKey()
	SetOfficialSiteBaseURL("")
	SetCommunityPushKey("")
	t.Cleanup(func() {
		SetOfficialSiteBaseURL(prevBase)
		SetCommunityPushKey(prevKey)
	})

	l := NewCommunityPushLogic()
	_, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "尚未配置社区推送")
}

func TestCommunityPushAuthErrorMapped(t *testing.T) {
	ctx, fake, db := setupCommunityPushTest(t)
	article := seedPushArticle(t, db)
	fake.setFail(401, "推送密钥无效或已吊销")

	l := NewCommunityPushLogic()
	_, err := l.PushToCommunity(ctx, TargetTypeArticle, article.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "推送密钥无效或已吊销")
}

func TestCommunityPushConfigRoundTrip(t *testing.T) {
	ctx, _, _ := setupCommunityPushTest(t)
	l := NewCommunityPushLogic()

	updated, err := l.UpdatePushConfig(ctx, &UpdatePushConfigRequest{
		BaseURL: "http://127.0.0.1:8090",
		SiteURL: "https://blog.example.com/",
		PushKey: "nbk_abcdef1234567890",
	})
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:8090/api/v1", updated.BaseURL)
	assert.Equal(t, "https://blog.example.com", updated.SiteURL)
	assert.True(t, updated.HasPushKey)
	assert.Equal(t, "nbk_abcdef12", updated.PushKeyPrefix)

	// 密钥留空 = 不修改
	updated2, err := l.UpdatePushConfig(ctx, &UpdatePushConfigRequest{BaseURL: "http://127.0.0.1:8090/api/v1", SiteURL: "https://blog.example.com"})
	require.NoError(t, err)
	assert.True(t, updated2.HasPushKey)
	assert.Equal(t, "nbk_abcdef12", updated2.PushKeyPrefix)
}

func TestCommunityPushVerifyCode(t *testing.T) {
	ctx, fake, _ := setupCommunityPushTest(t)
	l := NewCommunityPushLogic()

	t.Run("save with verify code submits to community", func(t *testing.T) {
		result, err := l.UpdatePushConfig(ctx, &UpdatePushConfigRequest{
			BaseURL:    fake.server.URL,
			SiteURL:    "https://blog.example.com",
			VerifyCode: "verify-code-123",
		})
		require.NoError(t, err)
		assert.Equal(t, "verified", result.VerifyStatus)
		assert.Contains(t, result.VerifyMessage, "alice")
		assert.Equal(t, "verify-code-123", result.VerifyCode)
		assert.Equal(t, "verify-code-123", fake.getLastVerifyCode())
	})

	t.Run("save without verify code skips submission", func(t *testing.T) {
		count := fake.getVerifyCount()
		result, err := l.UpdatePushConfig(ctx, &UpdatePushConfigRequest{
			BaseURL: fake.server.URL,
			SiteURL: "https://blog.example.com",
		})
		require.NoError(t, err)
		assert.Equal(t, "skipped", result.VerifyStatus)
		assert.Empty(t, result.VerifyMessage)
		assert.Equal(t, count, fake.getVerifyCount())
	})

	t.Run("community rejection surfaces message but keeps saved code", func(t *testing.T) {
		fake.setVerifyFail(400, "验证码无效，请在社区重新复制")
		result, err := l.UpdatePushConfig(ctx, &UpdatePushConfigRequest{
			BaseURL:    fake.server.URL,
			SiteURL:    "https://blog.example.com",
			VerifyCode: "bad-code",
		})
		require.NoError(t, err)
		assert.Equal(t, "failed", result.VerifyStatus)
		assert.Contains(t, result.VerifyMessage, "验证码无效")

		// 验证码仍保存，修正后重新保存即可重试
		config, err := l.GetPushConfig(ctx)
		require.NoError(t, err)
		assert.Equal(t, "bad-code", config.VerifyCode)
	})
}

// openRootPostgres 连接本地默认库（用于创建测试数据库）。
func openRootPostgres(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}
