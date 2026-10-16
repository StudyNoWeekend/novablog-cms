// Package novapush 提供与 NovaBlog 社区推送 API 对接的轻量客户端。
// 设计与 novablogapi 一致：无状态函数、凭据由调用方显式传入、统一解析官方 {code,message,data} 信封。
package novapush

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 15 * time.Second}

// apiPrefix 社区 API 版本前缀。
const apiPrefix = "/api/v1"

// AuthError 推送密钥无效或已吊销（社区返回 401）。
type AuthError struct {
	Message string
}

// Error 实现 error 接口。
func (e *AuthError) Error() string { return e.Message }

// APIError 社区返回的其他业务错误（参数错误、类型无效、限流 429 等）。
type APIError struct {
	Code    int
	Message string
}

// Error 实现 error 接口。
func (e *APIError) Error() string { return e.Message }

// NormalizeBaseURL 归一化社区地址：去尾部斜杠，未携带 /api/v1 前缀时自动补全。
func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("社区地址不能为空")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("社区地址必须为合法的 http(s) 地址")
	}
	raw = strings.TrimRight(raw, "/")
	if !strings.HasSuffix(raw, apiPrefix) {
		raw += apiPrefix
	}
	return raw, nil
}

// communityResp 社区统一响应 wrapper。
type communityResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// call 统一发起请求并解析社区响应。
func call(ctx context.Context, baseURL, method, path, pushKey string, body, out any) error {
	var reader io.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("编码请求体失败: %w", err)
		}
		reader = bytes.NewReader(bs)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if pushKey != "" {
		req.Header.Set("Authorization", "Bearer "+pushKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求社区服务失败: %w", err)
	}
	defer resp.Body.Close()

	var envelope communityResp
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("解析社区响应失败(HTTP %d): %w", resp.StatusCode, err)
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || envelope.Code == http.StatusUnauthorized:
		msg := envelope.Message
		if msg == "" {
			msg = "推送密钥无效或已吊销"
		}
		return &AuthError{Message: msg}
	case envelope.Code == http.StatusOK || envelope.Code == http.StatusCreated:
		if out != nil && len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, out); err != nil {
				return fmt.Errorf("解析社区数据失败: %w", err)
			}
		}
		return nil
	default:
		msg := envelope.Message
		if msg == "" {
			msg = resp.Status
		}
		return &APIError{Code: envelope.Code, Message: msg}
	}
}

// PingInfo ping 响应：确认密钥有效并回显博主信息。
type PingInfo struct {
	Username       string `json:"username"`
	BlogURL        string `json:"blog_url"`
	CredentialName string `json:"credential_name"`
}

// Ping 推送通道连通性自检。
func Ping(ctx context.Context, baseURL, pushKey string) (*PingInfo, error) {
	var out PingInfo
	if err := call(ctx, baseURL, http.MethodGet, "/push/ping", pushKey, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyBlogResult 站点归属验证结果。
type VerifyBlogResult struct {
	Verified bool   `json:"verified"`
	Username string `json:"username"` // 验证码所属的社区账号
}

// VerifyBlog 提交站点归属验证码（验证码本身即凭据，无需推送密钥）。
func VerifyBlog(ctx context.Context, baseURL, code string) (*VerifyBlogResult, error) {
	var out VerifyBlogResult
	body := map[string]string{"verify_code": code}
	if err := call(ctx, baseURL, http.MethodPost, "/verify/blog", "", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ArticleTypeInfo 社区内容类型字典条目。
type ArticleTypeInfo struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// Types 获取社区启用中的内容类型字典。
func Types(ctx context.Context, baseURL, pushKey string) ([]ArticleTypeInfo, error) {
	var out []ArticleTypeInfo
	if err := call(ctx, baseURL, http.MethodGet, "/push/types", pushKey, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertArticlePayload 推送文章请求体（与社区 POST /push/articles 契约一致）。
type UpsertArticlePayload struct {
	ExternalID      string         `json:"external_id"`
	Type            string         `json:"type"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary,omitempty"`
	ContentMarkdown string         `json:"content_markdown"`
	CoverImage      string         `json:"cover_image,omitempty"`
	Tags            []string       `json:"tags,omitempty"`
	SourceURL       string         `json:"source_url,omitempty"`
	PublishedAt     string         `json:"published_at,omitempty"` // RFC3339
	Status          string         `json:"status,omitempty"`       // published | hidden
	Extra           map[string]any `json:"extra,omitempty"`        // 类型专属结构化参数（与 CMS 内容结构同名）
}

// UpsertResult 推送结果。
type UpsertResult struct {
	ArticleID  uint   `json:"article_id"`
	ExternalID string `json:"external_id"`
	Result     string `json:"result"` // created | updated
}

// UpsertArticle 推送文章（创建与更新共用，按 external_id 幂等）。
func UpsertArticle(ctx context.Context, baseURL, pushKey string, payload *UpsertArticlePayload) (*UpsertResult, error) {
	var out UpsertResult
	if err := call(ctx, baseURL, http.MethodPost, "/push/articles", pushKey, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteResult 删除结果。
type DeleteResult struct {
	ExternalID string `json:"external_id"`
	Result     string `json:"result"` // deleted | not_found
}

// DeleteArticle 删除社区文章（幂等，not_found 亦返回 200）。
func DeleteArticle(ctx context.Context, baseURL, pushKey, externalID string) (*DeleteResult, error) {
	var out DeleteResult
	path := "/push/articles/" + url.PathEscape(externalID)
	if err := call(ctx, baseURL, http.MethodDelete, path, pushKey, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SyncItem 对账条目。
type SyncItem struct {
	ArticleID  uint   `json:"article_id"`
	ExternalID string `json:"external_id"`
	Title      string `json:"title"`
	Status     int    `json:"status"`
	UpdatedAt  string `json:"updated_at"`
}

// ListSync 回查社区同步状态（对账）。
func ListSync(ctx context.Context, baseURL, pushKey string, page, size int) ([]SyncItem, int64, error) {
	path := fmt.Sprintf("/push/articles?page=%d&size=%d", page, size)
	var envelope struct {
		Total int64      `json:"total"`
		Page  int        `json:"page"`
		Size  int        `json:"size"`
		Items []SyncItem `json:"items"`
	}
	if err := call(ctx, baseURL, http.MethodGet, path, pushKey, nil, &envelope); err != nil {
		return nil, 0, err
	}
	return envelope.Items, envelope.Total, nil
}
