// Package bingimages 提供基于 Bing 图片搜索异步接口的按关键词找图能力，
// 用于没有结构化图源的场景（如美食菜谱封面）。
//
// 异步接口返回 HTML 片段，图片直链藏在 m 属性 JSON 的 murl 字段里，正则提取即可。
package bingimages

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"novablog/pkg/proxyhttp"
)

const (
	asyncURL    = "https://cn.bing.com/images/async"
	fetchLimit  = 2 << 20 // 响应体读取上限 2MB
	readTimeout = 15 * time.Second

	defaultLimit = 12
	maxLimit     = 30
)

// 浏览器 UA：异步接口对无 UA 请求返回空内容。
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

// murlRe 从异步接口返回的 HTML 中提取图片直链（m 属性 JSON 的 murl 字段）。
var murlRe = regexp.MustCompile(`murl&quot;:&quot;(.*?)&quot;`)

// Image 搜索命中的图片结果。
type Image struct {
	Name string `json:"name"` // 网页图片无标题元数据，恒为空串
	URL  string `json:"url"`  // 原图直链
}

// Client Bing 图片搜索客户端。
type Client struct {
	// ProxyURL 可选 HTTP 代理；空=直连，代理不可用自动降级直连。
	ProxyURL string
}

// Search 按关键词搜索图片直链。
func (c *Client) Search(ctx context.Context, keyword string, limit int) ([]Image, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}

	endpoint := fmt.Sprintf("%s?q=%s&first=0&count=%d&mmasync=1", asyncURL, url.QueryEscape(keyword), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: readTimeout})
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Bing 图片失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, fetchLimit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bing 图片返回状态码 %d", resp.StatusCode)
	}

	out := make([]Image, 0, limit)
	seen := make(map[string]struct{})
	for _, m := range murlRe.FindAllStringSubmatch(string(body), -1) {
		u := html.UnescapeString(m[1])
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, Image{URL: u})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
