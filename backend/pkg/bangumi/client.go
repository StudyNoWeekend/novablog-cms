// Package bangumi 提供 Bangumi（bgm.tv）条目搜索能力，用于按名称查找条目封面（游戏、书籍等）。
//
// Bangumi API 国内可直连、无需鉴权，仅要求带意义的 User-Agent；
// 搜索接口处于 v0 版本，按官方要求附带 Confirm-Path 请求头。
package bangumi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"novablog/pkg/proxyhttp"
)

const (
	apiBase      = "https://api.bgm.tv"
	userAgent    = "novablog-admin/1.0 (https://github.com/czf/nova-blog-admin)"
	searchLimit  = 5 << 20 // 响应体读取上限 5MB
	readTimeout  = 20 * time.Second
	defaultLimit = 12
)

// Cover 搜索命中的封面结果。
type Cover struct {
	Name string `json:"name"` // 条目名（优先中文译名）
	URL  string `json:"url"`  // 封面图直链
}

// Client Bangumi 搜索客户端。
type Client struct {
	// ProxyURL 可选 HTTP 代理；空=直连，代理不可用自动降级直连。
	ProxyURL string
}

// subject 搜索结果条目中需要的字段。
type subject struct {
	Name   string `json:"name"`
	NameCn string `json:"name_cn"`
	Images struct {
		Large  string `json:"large"`
		Common string `json:"common"`
		Medium string `json:"medium"`
	} `json:"images"`
}

type searchResp struct {
	Data  []subject `json:"data"`
	Total int       `json:"total"`
}

// SearchSubjects 按关键词搜索条目，返回候选封面。
// subjectType 为 Bangumi 条目类型：4=游戏，1=书籍。
func (c *Client) SearchSubjects(ctx context.Context, keyword string, subjectType, limit int) ([]Cover, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}
	if limit <= 0 || limit > 30 {
		limit = defaultLimit
	}

	payload, err := json.Marshal(map[string]any{
		"keyword": keyword,
		// v0 搜索为全类型条目接口，需按类型过滤
		"filter": map[string]any{"type": []int{subjectType}},
	})
	if err != nil {
		return nil, err
	}
	// limit/offset 为查询参数（请求体中不生效）
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v0/search/subjects?limit=%d&offset=0", apiBase, limit), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Confirm-Path", "/v0/search/subjects")

	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: readTimeout})
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Bangumi 失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, searchLimit))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bangumi 返回状态码 %d", resp.StatusCode)
	}

	var parsed searchResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("解析 Bangumi 响应失败: %w", err)
	}
	covers := make([]Cover, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		name := item.NameCn
		if name == "" {
			name = item.Name
		}
		url := item.Images.Large
		if url == "" {
			url = item.Images.Common
		}
		if url == "" {
			url = item.Images.Medium
		}
		if name == "" || url == "" {
			continue
		}
		// 条目图床是 http 直链，统一升级为 https，避免前端页面混合内容告警
		covers = append(covers, Cover{Name: name, URL: strings.Replace(url, "http://", "https://", 1)})
	}
	return covers, nil
}
