package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"novablog/internal/cache"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BiliVideoPage B站视频分P信息
type BiliVideoPage struct {
	CID      int64
	Part     string // 分P标题
	Duration int
}

// BiliVideoInfo B站视频信息
type BiliVideoInfo struct {
	Title     string
	Pic       string
	Desc      string
	OwnerName string
	CID       int64 // 默认分P的cid
	Duration  int
	Pages     []BiliVideoPage
}

// ParseResult 解析结果
type ParseResult struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	CoverURL string `json:"cover_url"`
	Duration int    `json:"duration"`
	BVID     string `json:"bvid"`
	CID      int64  `json:"cid"`
}

// ParseTask 解析任务
type ParseTask struct {
	ID        string        `json:"id"`
	URL       string        `json:"url"`
	Status    string        `json:"status"`
	Results   []ParseResult `json:"results"`
	Error     string        `json:"error"`
	CreatedAt time.Time     `json:"created_at"`
}

// biliVideoInfoResp B站视频信息 API 响应
type biliVideoInfoResp struct {
	Code int `json:"code"`
	Data struct {
		Title    string `json:"title"`
		Pic      string `json:"pic"`
		Desc     string `json:"desc"`
		CID      int64  `json:"cid"`
		Duration int    `json:"duration"`
		Owner    struct {
			Name string `json:"name"`
		} `json:"owner"`
		Pages []struct {
			CID      int64  `json:"cid"`
			Part     string `json:"part"`
			Duration int    `json:"duration"`
		} `json:"pages"`
	} `json:"data"`
}

// biliPlayurlResp B站播放地址 API 响应
type biliPlayurlResp struct {
	Code int `json:"code"`
	Data struct {
		Dash struct {
			Audio []struct {
				ID        int    `json:"id"`
				BaseURL   string `json:"baseUrl"`
				Bandwidth int    `json:"bandwidth"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}

var biliHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// newBiliRequest 创建带有 B 站所需请求头的 GET 请求
func newBiliRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")
	return req, nil
}

var bvidRegex = regexp.MustCompile(`BV[a-zA-Z0-9]{10}`)

// ExtractBVID 从多种 B 站 URL 格式中提取 BV 号
func ExtractBVID(url string) (string, error) {
	// 短链暂不支持解析
	if matched, _ := regexp.MatchString(`^https?://b23\.tv/`, url); matched {
		return "", fmt.Errorf("暂不支持 b23.tv 短链解析，请提供完整视频地址")
	}

	bvid := bvidRegex.FindString(url)
	if bvid == "" {
		return "", fmt.Errorf("无法从 URL 中提取 BV 号: %s", url)
	}
	return bvid, nil
}

// FetchVideoInfo 获取 B 站视频信息
func FetchVideoInfo(ctx context.Context, bvid string) (*BiliVideoInfo, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)

	req, err := newBiliRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := biliHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求视频信息失败: %w", err)
	}
	defer resp.Body.Close()

	var body biliVideoInfoResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("解析视频信息响应失败: %w", err)
	}

	if body.Code != 0 {
		return nil, fmt.Errorf("获取视频信息失败, code: %d", body.Code)
	}

	info := &BiliVideoInfo{
		Title:     body.Data.Title,
		Pic:       body.Data.Pic,
		Desc:      body.Data.Desc,
		OwnerName: body.Data.Owner.Name,
		CID:       body.Data.CID,
		Duration:  body.Data.Duration,
	}
	for _, p := range body.Data.Pages {
		info.Pages = append(info.Pages, BiliVideoPage{
			CID:      p.CID,
			Part:     p.Part,
			Duration: p.Duration,
		})
	}
	return info, nil
}

// FetchAudioURL 获取 B 站音频流地址（选择 bandwidth 最大的条目）
func FetchAudioURL(ctx context.Context, bvid string, cid int64) (string, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/player/playurl?bvid=%s&cid=%d&fnval=16&fnver=0", bvid, cid)

	req, err := newBiliRequest(ctx, url)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := biliHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求播放地址失败: %w", err)
	}
	defer resp.Body.Close()

	var body biliPlayurlResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("解析播放地址响应失败: %w", err)
	}

	if body.Code != 0 {
		return "", fmt.Errorf("获取播放地址失败, code: %d", body.Code)
	}

	if len(body.Data.Dash.Audio) == 0 {
		return "", fmt.Errorf("未找到可用的音频流")
	}

	var bestURL string
	var bestBandwidth int
	for _, audio := range body.Data.Dash.Audio {
		if audio.Bandwidth > bestBandwidth {
			bestBandwidth = audio.Bandwidth
			bestURL = audio.BaseURL
		}
	}

	if bestURL == "" {
		return "", fmt.Errorf("未找到可用的音频流")
	}

	return bestURL, nil
}

// parseTaskStore 内存任务存储
var parseTaskStore sync.Map

// biliParseCacheTTL B站解析结果缓存有效期
const biliParseCacheTTL = 7 * 24 * time.Hour

// biliParseCacheKey 生成 Redis 缓存 key
func biliParseCacheKey(bvid string) string {
	return "bili:parse:" + bvid
}

// getCachedParseResults 从 Redis 获取缓存的解析结果
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

// setCachedParseResults 将解析结果存入 Redis
func setCachedParseResults(ctx context.Context, bvid string, results []ParseResult) {
	if cache.RedisClient == nil || len(results) == 0 {
		return
	}
	data, err := json.Marshal(results)
	if err != nil {
		return
	}
	cache.RedisClient.Set(ctx, biliParseCacheKey(bvid), data, biliParseCacheTTL)
}

// SaveParseTask 保存解析任务，并在 10 分钟后自动清理
func SaveParseTask(task *ParseTask) {
	parseTaskStore.Store(task.ID, task)
	go func() {
		time.Sleep(10 * time.Minute)
		parseTaskStore.Delete(task.ID)
	}()
}

// GetParseTask 根据任务 ID 获取解析任务
func GetParseTask(taskID string) (*ParseTask, bool) {
	val, ok := parseTaskStore.Load(taskID)
	if !ok {
		return nil, false
	}
	return val.(*ParseTask), true
}

// StartParseTask 启动解析任务，返回 task_id
func StartParseTask(url string) (string, error) {
	bvid, err := ExtractBVID(url)
	if err != nil {
		return "", err
	}

	// 先查 Redis 缓存，命中则直接返回成功任务
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
		info, err := FetchVideoInfo(ctx, bvid)
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
					Title:    page.Part, // 分P标题
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
