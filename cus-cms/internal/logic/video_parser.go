package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// ParsedVideoMeta 解析到的视频元信息
type ParsedVideoMeta struct {
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
	Description string `json:"description"`
}

var parseHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// ParseVideoMeta 根据平台和 URL 解析视频元信息
func ParseVideoMeta(ctx context.Context, platform, url string) (*ParsedVideoMeta, error) {
	switch platform {
	case "bilibili":
		return parseBilibili(ctx, url)
	case "youtube":
		return parseOEmbed(ctx, "https://www.youtube.com/oembed?url="+url+"&format=json")
	case "vimeo":
		return parseOEmbed(ctx, "https://vimeo.com/api/oembed.json?url="+url)
	default:
		return parseOpenGraph(ctx, url)
	}
}

// parseBilibili 解析 B 站视频元信息
func parseBilibili(ctx context.Context, url string) (*ParsedVideoMeta, error) {
	bvid, err := ExtractBVID(url)
	if err != nil {
		return nil, err
	}
	info, err := FetchVideoInfo(ctx, bvid)
	if err != nil {
		return nil, err
	}
	return &ParsedVideoMeta{
		Title:       info.Title,
		CoverURL:    info.Pic,
		Description: info.Desc,
	}, nil
}

// oEmbedResp oEmbed 标准响应结构体
type oEmbedResp struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Description  string `json:"description"`
}

// parseOEmbed 通过 oEmbed 接口解析视频元信息
func parseOEmbed(ctx context.Context, oembedURL string) (*ParsedVideoMeta, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", oembedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := parseHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 oEmbed 接口失败: %w", err)
	}
	defer resp.Body.Close()

	var body oEmbedResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("解析 oEmbed 响应失败: %w", err)
	}

	return &ParsedVideoMeta{
		Title:       body.Title,
		CoverURL:    body.ThumbnailURL,
		Description: body.Description,
	}, nil
}

// parseOpenGraph 通过 OpenGraph 协议解析视频元信息
func parseOpenGraph(ctx context.Context, url string) (*ParsedVideoMeta, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := parseHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求页面失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取页面内容失败: %w", err)
	}

	html := string(body)
	meta := &ParsedVideoMeta{}
	meta.Title = extractMetaContent(html, "og:title")
	meta.CoverURL = extractMetaContent(html, "og:image")
	meta.Description = extractMetaContent(html, "og:description")

	if meta.Title == "" && meta.CoverURL == "" && meta.Description == "" {
		return nil, fmt.Errorf("无法从该链接解析到视频信息")
	}
	return meta, nil
}

// extractMetaContent 从 HTML 中提取指定 property 的 meta 标签 content 值
func extractMetaContent(html, property string) string {
	// 匹配 property 在前和 content 在前的两种情况
	pattern := fmt.Sprintf(`<meta\s+(?:property|name)=["']%s["']\s+content=["']([^"']*)["']|<meta\s+content=["']([^"']*)["']\s+(?:property|name)=["']%s["']`, property, property)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		if matches[1] != "" {
			return matches[1]
		}
		if len(matches) > 2 && matches[2] != "" {
			return matches[2]
		}
	}
	return ""
}
