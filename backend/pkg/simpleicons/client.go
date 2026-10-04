// Package simpleicons 提供 Simple Icons 品牌图标的按名称搜索能力。
//
// Simple Icons 没有公开的搜索 API，这里采取"拉取一次图标数据集 + 内存缓存 + 本地匹配"：
// 数据集从 jsDelivr / unpkg / GitHub raw 依次回退获取，缓存 24 小时。
// 图标本体不下载，直接引用 https://cdn.simpleicons.org/{slug}（默认品牌色 SVG）。
package simpleicons

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"novablog/pkg/proxyhttp"
)

const (
	// dataTTL 数据集缓存时长：图标库低频更新，24h 足够新鲜。
	dataTTL = 24 * time.Hour
	// fetchTimeout 单次数据集拉取超时（数据集约 3MB，放宽一些）。
	fetchTimeout = 30 * time.Second
	// fetchLimit 响应体读取上限 10MB，防御异常来源。
	fetchLimit = 10 << 20
	// minDatasetSize 数据集最小条目数：低于该值视为来源返回了异常内容。
	minDatasetSize = 500
	// defaultLimit 默认返回条数。
	defaultLimit = 24
)

// datasetSources 数据集来源，按国内可达性依次回退。
var datasetSources = []string{
	"https://cdn.jsdelivr.net/npm/simple-icons@latest/_data/simple-icons.json",
	"https://fastly.jsdelivr.net/npm/simple-icons@latest/_data/simple-icons.json",
	"https://unpkg.com/simple-icons/_data/simple-icons.json",
	"https://raw.githubusercontent.com/simple-icons/simple-icons/develop/_data/simple-icons.json",
}

// Icon 搜索命中的图标结果。
type Icon struct {
	Title string `json:"title"` // 图标名称，如 "Vue.js"
	Slug  string `json:"slug"`  // 图标 slug，如 "vuedotjs"
	Hex   string `json:"hex"`   // 品牌色，如 "4FC08D"
	URL   string `json:"url"`   // 图标直链（默认品牌色 SVG）
}

// simpleIcon 数据集条目（只取需要的字段；slug 在数据集中可能缺省，需按标题推导）。
type simpleIcon struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Hex   string `json:"hex"`
}

// slug 返回图标 slug：优先用数据集自带值，否则按官方规则从标题推导。
func (i simpleIcon) slug() string {
	if i.Slug != "" {
		return i.Slug
	}
	return slugify(i.Title)
}

// slugify 按官方规则从标题推导 slug：小写、"." 转 "dot"、去掉其余非字母数字字符。
// 少数含特殊符号的标题有官方特例，单独维护。
func slugify(title string) string {
	switch title {
	case "C++":
		return "cplusplus"
	case "C#":
		return "csharp"
	case "F#":
		return "fsharp"
	}
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, ".", "dot")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Client Simple Icons 搜索客户端。
type Client struct {
	// ProxyURL 可选 HTTP 代理；空=直连，代理不可用自动降级直连。
	ProxyURL string

	mu      sync.Mutex
	icons   []simpleIcon
	fetched time.Time
}

// Search 按关键词搜索图标，返回按匹配度排序的结果（标题前缀 > slug 前缀 > 标题包含 > slug 包含）。
func (c *Client) Search(ctx context.Context, keyword string, limit int) ([]Icon, error) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}
	if limit <= 0 || limit > 50 {
		limit = defaultLimit
	}

	icons, err := c.dataset(ctx)
	if err != nil {
		return nil, err
	}

	type hit struct {
		icon simpleIcon
		rank int
	}
	var hits []hit
	for _, ic := range icons {
		title := strings.ToLower(ic.Title)
		slug := ic.slug()
		rank := -1
		switch {
		case strings.HasPrefix(title, keyword):
			rank = 0
		case strings.HasPrefix(slug, keyword):
			rank = 1
		case strings.Contains(title, keyword):
			rank = 2
		case strings.Contains(slug, keyword):
			rank = 3
		}
		if rank >= 0 {
			hits = append(hits, hit{icon: ic, rank: rank})
		}
	}
	sort.Slice(hits, func(a, b int) bool {
		if hits[a].rank != hits[b].rank {
			return hits[a].rank < hits[b].rank
		}
		return len(hits[a].icon.Title) < len(hits[b].icon.Title)
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}

	out := make([]Icon, 0, len(hits))
	for _, h := range hits {
		slug := h.icon.slug()
		out = append(out, Icon{
			Title: h.icon.Title,
			Slug:  slug,
			Hex:   h.icon.Hex,
			URL:   "https://cdn.simpleicons.org/" + slug,
		})
	}
	return out, nil
}

// dataset 获取图标数据集：带 TTL 缓存，未命中时按来源顺序回退拉取。
func (c *Client) dataset(ctx context.Context) ([]simpleIcon, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.icons != nil && time.Since(c.fetched) < dataTTL {
		return c.icons, nil
	}

	var lastErr error
	for _, src := range datasetSources {
		list, err := c.fetchDataset(ctx, src)
		if err != nil {
			lastErr = err
			continue
		}
		if len(list) < minDatasetSize {
			lastErr = fmt.Errorf("来源 %s 返回条目数异常（%d）", src, len(list))
			continue
		}
		c.icons = list
		c.fetched = time.Now()
		return c.icons, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("无可用数据集来源")
	}
	return nil, fmt.Errorf("获取 Simple Icons 数据集失败: %w", lastErr)
}

// fetchDataset 从单一来源拉取并解析数据集。
func (c *Client) fetchDataset(ctx context.Context, src string) ([]simpleIcon, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return nil, err
	}
	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: fetchTimeout})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("状态码 %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, fetchLimit))
	if err != nil {
		return nil, err
	}
	var list []simpleIcon
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return list, nil
}
