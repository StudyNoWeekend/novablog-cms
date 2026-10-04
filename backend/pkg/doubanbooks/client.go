// Package doubanbooks 提供豆瓣读书的按名称搜索能力。
//
// 使用豆瓣站内搜索框的 suggest JSON 接口（book.douban.com/j/subject_suggest），
// 国内可直连、无需鉴权，但对无 Cookie 的图片请求有指纹级反爬（418）：
// 客户端维护 CookieJar，先访问主站拿到 bid Cookie，后续下载封面图即可通过。
package doubanbooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"novablog/pkg/proxyhttp"
)

const (
	suggestURL  = "https://book.douban.com/j/subject_suggest"
	siteBase    = "https://book.douban.com"
	searchLimit = 10 << 20 // 封面下载上限 10MB
	readTimeout = 20 * time.Second

	defaultLimit = 12
	maxLimit     = 20
)

// 浏览器 UA：豆瓣对无 UA / 裸 UA 请求直接拒绝。
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// Cover 搜索命中的封面结果。
type Cover struct {
	Name  string `json:"name"`  // 书名（带作者时附加在书名后）
	URL   string `json:"url"`   // 封面图直链（大图）
	Year  string `json:"year"`  // 出版年份
	Extra string `json:"extra"` // 条目页地址（备用）
}

// suggestItem suggest 接口返回条目。
type suggestItem struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Pic        string `json:"pic"`
	AuthorName string `json:"author_name"`
	Year       string `json:"year"`
	Type       string `json:"type"`
}

// Client 豆瓣图书搜索客户端。CookieJar 在多次请求间保留 bid Cookie。
type Client struct {
	// ProxyURL 可选 HTTP 代理；空=直连，代理不可用自动降级直连。
	ProxyURL string

	jar http.CookieJar
}

func (c *Client) initJar() {
	if c.jar == nil {
		c.jar, _ = cookiejar.New(nil)
	}
}

func (c *Client) httpClient() *http.Client {
	c.initJar()
	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: readTimeout})
	client.Jar = c.jar
	return client
}

// ensureBid 确保已持有豆瓣 bid Cookie（图片 CDN 校验的关键）。
func (c *Client) ensureBid(ctx context.Context) error {
	c.initJar()
	site, err := url.Parse(siteBase)
	if err != nil {
		return err
	}
	for _, ck := range c.jar.Cookies(site) {
		if ck.Name == "bid" && ck.Value != "" {
			return nil
		}
	}
	return c.warmup(ctx)
}

// rewarm 重置 CookieJar（换取全新 bid）并访问主站预热。
// 豆瓣对已持有 bid 的请求不会下发新 Cookie，想换 bid 必须换全新 Jar。
func (c *Client) rewarm(ctx context.Context) error {
	c.jar, _ = cookiejar.New(nil)
	return c.warmup(ctx)
}

// warmup 访问一次主站触发 Set-Cookie bid。
func (c *Client) warmup(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, siteBase+"/", nil)
	if err != nil {
		return err
	}
	c.setCommonHeaders(req)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("访问豆瓣主站失败: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
}

// Search 按书名/作者搜索图书，返回候选封面（大图）。
func (c *Client) Search(ctx context.Context, keyword string, limit int) ([]Cover, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if err := c.ensureBid(ctx); err != nil {
		return nil, err
	}

	endpoint := suggestURL + "?q=" + url.QueryEscape(keyword)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Referer", siteBase+"/")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求豆瓣读书失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("豆瓣读书返回状态码 %d", resp.StatusCode)
	}

	var items []suggestItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("解析豆瓣读书响应失败: %w", err)
	}
	out := make([]Cover, 0, len(items))
	for _, item := range items {
		if item.Type != "b" || item.Pic == "" {
			continue
		}
		name := item.Title
		if item.AuthorName != "" {
			name = fmt.Sprintf("%s（%s）", item.Title, item.AuthorName)
		}
		out = append(out, Cover{
			Name:  name,
			URL:   strings.Replace(item.Pic, "/view/subject/s/public/", "/view/subject/l/public/", 1),
			Year:  item.Year,
			Extra: item.URL,
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// DownloadCover 下载豆瓣封面图字节，返回字节与 Content-Type。
// 豆瓣反爬是概率性的（命中时返回 200 + HTML 验证页），按以下链条逐级兜底：
//  1. 无 Cookie 直连（多数情况可用）
//  2. 全新 CookieJar（换新 bid）+ 大图
//  3. 全新 CookieJar + 小图（s）
//  4. weserv 公共图片代理
//
// 注意：豆瓣对已持有 bid 的请求不会下发新 Cookie，因此重试必须换全新会话，
// 只重复请求主站是无效的。
func (c *Client) DownloadCover(ctx context.Context, imgURL string) ([]byte, string, error) {
	if !strings.HasPrefix(imgURL, "http://") && !strings.HasPrefix(imgURL, "https://") {
		return nil, "", fmt.Errorf("图片地址不合法")
	}

	var data []byte
	var contentType string
	var lastErr error

	// 1. 无 Cookie 直连
	if data, contentType, lastErr = c.fetchPlain(ctx, imgURL); lastErr == nil {
		return data, contentType, nil
	}

	// 2. 全新会话（新 bid）+ 原始地址
	if err := c.rewarm(ctx); err == nil {
		if data, contentType, lastErr = c.doDownload(ctx, imgURL); lastErr == nil {
			return data, contentType, nil
		}
	}

	// 3. 全新会话 + 小图变体
	if small := strings.Replace(imgURL, "/l/public/", "/s/public/", 1); small != imgURL {
		if err := c.rewarm(ctx); err == nil {
			if data, contentType, lastErr = c.doDownload(ctx, small); lastErr == nil {
				return data, contentType, nil
			}
		}
	}

	// 4. weserv 公共图片代理兜底
	if data, contentType, lastErr = c.fetchViaWeserv(ctx, imgURL); lastErr == nil {
		return data, contentType, nil
	}

	return nil, "", fmt.Errorf("豆瓣封面下载被拦截，请稍后重试: %w", lastErr)
}

// fetchPlain 无 Cookie 直连下载（最简单的路径，多数情况即可命中）。
func (c *Client) fetchPlain(ctx context.Context, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, "", err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Referer", siteBase+"/")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	client := proxyhttp.NewResilient(proxyhttp.Config{Timeout: readTimeout})
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("下载豆瓣封面失败: %w", err)
	}
	defer resp.Body.Close()
	return readImage(resp)
}

// fetchViaWeserv 经 weserv 公共图片代理下载（其服务端 IP 不受豆瓣反爬影响）。
func (c *Client) fetchViaWeserv(ctx context.Context, imgURL string) ([]byte, string, error) {
	stripped := strings.TrimPrefix(strings.TrimPrefix(imgURL, "https://"), "http://")
	proxyedImage := "https://images.weserv.nl/?url=" + url.QueryEscape(stripped)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, proxyedImage, nil)
	if err != nil {
		return nil, "", err
	}
	c.setCommonHeaders(req)

	client := proxyhttp.NewResilient(proxyhttp.Config{Timeout: readTimeout})
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("weserv 代理下载失败: %w", err)
	}
	defer resp.Body.Close()
	return readImage(resp)
}

// doDownload 单次尝试下载封面（走 CookieJar 会话）。
func (c *Client) doDownload(ctx context.Context, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, "", err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Referer", siteBase+"/")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("下载豆瓣封面失败: %w", err)
	}
	defer resp.Body.Close()
	return readImage(resp)
}

// readImage 校验响应为图片并读取字节。
func readImage(resp *http.Response) ([]byte, string, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("状态码 %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[:idx]
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("返回了非图片内容（Content-Type: %s，疑似反爬验证页）", contentType)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, searchLimit))
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}
