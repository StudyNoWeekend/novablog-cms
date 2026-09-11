package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// musicCacheKeyPrefix 歌曲音频 URL 缓存 key 前缀。
// 完整 key 格式：music:audio:<song_id>
const musicCacheKeyPrefix = "music:audio:"

// audioURLCacheTTL 兜底过期时间。
//
// B 站 /x/player/playurl 返回的 CDN URL 中通常带 expire 查询参数，
// 真实过期时间由 ExtractExpireFromURL 解析。解析失败时使用该兜底值。
const audioURLCacheTTL = 2 * time.Hour

// MusicCache 音乐相关缓存结构体。
type MusicCache struct {
	client *redis.Client
}

// NewMusicCache 创建 MusicCache 实例。
func NewMusicCache() *MusicCache {
	return &MusicCache{client: RedisClient}
}

// musicCacheKey 生成音频 URL 缓存的 Redis key。
func musicCacheKey(songID string) string {
	return musicCacheKeyPrefix + songID
}

// CachedAudioURL 缓存中的音频 URL 与过期信息。
// ExpireAt 是 B 站 URL 中 expire 参数对应的时间点（用于故障排查与日志）。
type CachedAudioURL struct {
	URL      string `json:"url"`
	ExpireAt int64  `json:"expire_at"`
}

// GetAudioURL 读取缓存的音频 URL，缓存不存在或读取失败时返回 (nil, nil)。
func (c *MusicCache) GetAudioURL(ctx context.Context, songID string) (*CachedAudioURL, error) {
	if c.client == nil {
		return nil, nil
	}
	data, err := c.client.Get(ctx, musicCacheKey(songID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取音频 URL 缓存失败: %w", err)
	}
	var v CachedAudioURL
	if err := json.Unmarshal(data, &v); err != nil {
		// 缓存数据损坏，删除以便下次重新拉取。
		_ = c.client.Del(ctx, musicCacheKey(songID)).Err()
		return nil, nil
	}
	return &v, nil
}

// SetAudioURL 写入音频 URL 缓存。
//
// 过期时间策略：
//  1. 优先从 URL 的 expire 查询参数解析出 B 站 CDN 真实失效时间；
//  2. 解析失败或已过期时使用 audioURLCacheTTL 兜底；
//  3. 写缓存时同时设 Redis TTL，避免冷数据长期驻留。
func (c *MusicCache) SetAudioURL(ctx context.Context, songID, audioURL string) error {
	if c.client == nil {
		return nil
	}

	expireAt, ttl := resolveExpire(audioURL)
	v := CachedAudioURL{URL: audioURL, ExpireAt: expireAt}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("序列化音频 URL 缓存失败: %w", err)
	}
	return c.client.Set(ctx, musicCacheKey(songID), data, ttl).Err()
}

// InvalidateAudioURL 主动失效缓存（歌曲被更新/删除时调用）。
func (c *MusicCache) InvalidateAudioURL(ctx context.Context, songID string) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, musicCacheKey(songID)).Err()
}

// resolveExpire 从 B 站 CDN URL 的 expire 查询参数解析真实失效时间。
// 返回 (expireAtUnix, ttl)；若 URL 无 expire 或解析失败，返回 (0, audioURLCacheTTL)。
//
// B 站 CDN URL 示例：https://...bilivideo.com/.../...m4s?expires=...
// 实际字段名依接口/cdn 而定，本实现同时兼容 `expire` 与 `expires` 两种命名。
func resolveExpire(rawURL string) (int64, time.Duration) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, audioURLCacheTTL
	}
	q := u.Query()
	expireStr := q.Get("expire")
	if expireStr == "" {
		expireStr = q.Get("expires")
	}
	if expireStr == "" {
		return 0, audioURLCacheTTL
	}
	expireUnix, err := strconv.ParseInt(expireStr, 10, 64)
	if err != nil || expireUnix <= 0 {
		return 0, audioURLCacheTTL
	}
	expireAt := time.Unix(expireUnix, 0)
	ttl := time.Until(expireAt)
	// URL 已过期或 TTL 不足 1 分钟：使用兜底 TTL。
	if ttl < time.Minute {
		return expireUnix, audioURLCacheTTL
	}
	// 留出 5% 余量，避免缓存边界与 CDN 边界同时失效。
	safetyMargin := time.Duration(float64(ttl) * 0.05)
	safetyMargin = min(safetyMargin, 30*time.Second)
	return expireUnix, ttl - safetyMargin
}
