package logic

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"novablog/internal/cache"
	"novablog/pkg/bilibili"

	"github.com/google/uuid"
)

// ParseResult 解析结果。
type ParseResult struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	CoverURL string `json:"cover_url"`
	Duration int    `json:"duration"`
	BVID     string `json:"bvid"`
	CID      int64  `json:"cid"`
}

// ParseTask 解析任务。
type ParseTask struct {
	ID        string        `json:"id"`
	URL       string        `json:"url"`
	Status    string        `json:"status"`
	Results   []ParseResult `json:"results"`
	Error     string        `json:"error"`
	CreatedAt time.Time     `json:"created_at"`
}

// parseTaskStore 内存任务存储（短生命周期任务，用 sync.Map 即可）。
var parseTaskStore sync.Map

// biliParseCacheTTL 解析结果在 Redis 中的有效期。
const biliParseCacheTTL = 7 * 24 * time.Hour

// biliParseCacheKey 生成 Redis 缓存 key。
func biliParseCacheKey(bvid string) string {
	return "bili:parse:" + bvid
}

// getCachedParseResults 从 Redis 获取缓存的解析结果。
func getCachedParseResults(ctx context.Context, bvid string) []ParseResult {
	if cache.RedisClient == nil {
		return nil
	}
	data, err := cache.RedisClient.Get(ctx, biliParseCacheKey(bvid)).Bytes()
	if err != nil || len(data) == 0 {
		return nil
	}
	var results []ParseResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil
	}
	return results
}

// setCachedParseResults 将解析结果存入 Redis。
func setCachedParseResults(ctx context.Context, bvid string, results []ParseResult) {
	if cache.RedisClient == nil || len(results) == 0 {
		return
	}
	data, err := json.Marshal(results)
	if err != nil {
		return
	}
	_ = cache.RedisClient.Set(ctx, biliParseCacheKey(bvid), data, biliParseCacheTTL).Err()
}

// SaveParseTask 保存解析任务，10 分钟后自动清理。
func SaveParseTask(task *ParseTask) {
	parseTaskStore.Store(task.ID, task)
	go func() {
		time.Sleep(10 * time.Minute)
		parseTaskStore.Delete(task.ID)
	}()
}

// GetParseTask 根据任务 ID 获取解析任务。
func GetParseTask(taskID string) (*ParseTask, bool) {
	val, ok := parseTaskStore.Load(taskID)
	if !ok {
		return nil, false
	}
	return val.(*ParseTask), true
}

// StartParseTask 启动解析任务，返回 task_id。
//
// 流程：
//  1. 从 URL 提取 BV 号；
//  2. 先查 Redis 解析结果缓存；命中则直接返回成功任务；
//  3. 否则异步调用 B 站接口解析，结果同时回填任务状态与 Redis 缓存。
func StartParseTask(url string) (string, error) {
	bvid, err := bilibili.ExtractBVID(url)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	if cached := getCachedParseResults(ctx, bvid); cached != nil {
		task := &ParseTask{
			ID:        uuid.New().String(),
			URL:       url,
			Status:    "success",
			Results:   cached,
			CreatedAt: time.Now(),
		}
		SaveParseTask(task)
		return task.ID, nil
	}

	task := &ParseTask{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	SaveParseTask(task)

	go func(taskID, bvid string) {
		ctx := context.Background()

		// 更新状态为 processing
		if t, ok := GetParseTask(taskID); ok {
			t.Status = "processing"
		}

		// 获取视频信息
		info, err := bilibili.FetchVideoInfo(ctx, bvid)
		if err != nil {
			if t, ok := GetParseTask(taskID); ok {
				t.Status = "failed"
				t.Error = err.Error()
			}
			return
		}

		var results []ParseResult
		if len(info.Pages) > 1 {
			// 多分P：每个分P一个结果
			for _, page := range info.Pages {
				results = append(results, ParseResult{
					Title:    page.Part,
					Artist:   info.OwnerName,
					CoverURL: info.Pic,
					Duration: page.Duration,
					BVID:     bvid,
					CID:      page.CID,
				})
			}
		} else {
			// 单分P：使用视频标题
			results = append(results, ParseResult{
				Title:    info.Title,
				Artist:   info.OwnerName,
				CoverURL: info.Pic,
				Duration: info.Duration,
				BVID:     bvid,
				CID:      info.CID,
			})
		}

		// 存入 Redis 缓存（7天）
		setCachedParseResults(ctx, bvid, results)

		// 成功，填充结果
		if t, ok := GetParseTask(taskID); ok {
			t.Status = "success"
			t.Results = results
		}
	}(task.ID, bvid)

	return task.ID, nil
}
