// Package novablogapi 提供 NovaBlog 官方主题市场 API 的 HTTP 客户端。
//
// 官方接口契约见 novablog 仓库 docs/NovaBlog接口文档.md（Base URL 形如 http://novablogapi.ditancafebar.cn/api/v1）。
// 所有方法均为无状态纯转发：官方地址与 Token 由调用方显式传入，
// 本包不缓存任何凭据，登录态管理由前端负责、错误映射由 logic 层编排。
package novablogapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// client 公用 HTTP 客户端，统一设置超时。
var client = &http.Client{Timeout: 15 * time.Second}

// downloadResolveClient 不跟随重定向的客户端，专用于解析官方代理下载地址的 302 Location。
var downloadResolveClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// apiPrefix 官方 API 版本前缀。
const apiPrefix = "/api/v1"

// officialDownloadPathPattern 官方代理下载地址的路径段：{apiPrefix}/themes/{id}/download。
var officialDownloadPathPattern = regexp.MustCompile(`^` + apiPrefix + `/themes/\d+/download/?$`)

// AuthError 官方返回 401：主题接口场景表示登录已失效，登录/刷新场景表示凭据错误。
type AuthError struct {
	Message string // 官方返回的错误描述
}

// Error 实现 error 接口。
func (e *AuthError) Error() string { return e.Message }

// APIError 官方返回的其他业务错误（参数错误、资源不存在等）。
type APIError struct {
	Code    int    // 官方业务码
	Message string // 官方错误描述
}

// Error 实现 error 接口。
func (e *APIError) Error() string { return e.Message }

// NormalizeBaseURL 归一化官方地址：去尾部斜杠，未携带 /api/v1 前缀时自动补全。
// 允许输入 http://novablogapi.ditancafebar.cn 或 http://novablogapi.ditancafebar.cn/api/v1 两种形式。
func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("官方地址不能为空")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("官方地址必须为合法的 http(s) 地址")
	}
	raw = strings.TrimRight(raw, "/")
	if !strings.HasSuffix(raw, apiPrefix) {
		raw += apiPrefix
	}
	return raw, nil
}

// officialResp 官方统一响应 wrapper。
type officialResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// call 统一发起请求并解析官方响应。
func call(ctx context.Context, baseURL, method, path, token string, body, out any) error {
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求官方服务失败: %w", err)
	}
	defer resp.Body.Close()

	var envelope officialResp
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("解析官方响应失败(HTTP %d): %w", resp.StatusCode, err)
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || envelope.Code == http.StatusUnauthorized:
		msg := envelope.Message
		if msg == "" {
			msg = "官方账号登录已失效"
		}
		return &AuthError{Message: msg}
	case envelope.Code == http.StatusOK || envelope.Code == http.StatusCreated:
		if out != nil && len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, out); err != nil {
				return fmt.Errorf("解析官方数据失败: %w", err)
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

// TokenPair 官方 Token。
type TokenPair struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// UserInfo 官方用户信息（登录接口返回）。
type UserInfo struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	Bio          string `json:"bio"`
	BlogURL      string `json:"blog_url"`
	BlogVerified bool   `json:"blog_verified"`
	Role         string `json:"role"`
	CreatedAt    string `json:"created_at"`
}

// loginResp 官方登录响应 data。
type loginResp struct {
	Token TokenPair `json:"token"`
	User  UserInfo  `json:"user"`
}

// ThemeQuery 主题列表查询参数，映射官方 GET /themes 的 query。
type ThemeQuery struct {
	Page     int
	Size     int
	Type     string
	Styles   []string
	Features []string
	Price    string
	Search   string
	Sort     string
}

// ThemeItem 主题条目，字段与官方 ThemeItem 一致。
type ThemeItem struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	AuthorID    string   `json:"author_id"`
	Author      string   `json:"author"`
	Price       string   `json:"price"`
	PriceAmount float64  `json:"price_amount"`
	Type        string   `json:"type"`
	Styles      []string `json:"styles"`
	Features    []string `json:"features"`
	Preview     string   `json:"preview"`
	Version     string   `json:"version"`
	Downloads   int64    `json:"downloads"`
	Likes       int64    `json:"likes"`
	Rating      float64  `json:"rating"`
	Status      int      `json:"status"`
	CreatedAt   string   `json:"created_at"`
}

// ThemeList 主题分页列表（官方原始分页结构）。
type ThemeList struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Items []ThemeItem `json:"items"`
}

// ThemeDetail 主题详情（官方 ThemeItem 全部字段 + 登录态个人状态）。
type ThemeDetail struct {
	ThemeItem
	DownloadURL string  `json:"download_url"`
	Liked       bool    `json:"liked"`
	UserRating  float64 `json:"user_rating"`
}

// ThemeReleaseItem 主题版本历史条目（官方自 GitHub Release 同步的版本与更新日志）。
type ThemeReleaseItem struct {
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Notes       string `json:"notes"`
	PublishedAt string `json:"published_at"`
}

// ThemeStats 市场统计。
type ThemeStats struct {
	Total     int64 `json:"total"`
	Authors   int64 `json:"authors"`
	Downloads int64 `json:"downloads"`
}

// HotTag 热门风格标签。
type HotTag struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// LikeResult 点赞/取消点赞结果。
type LikeResult struct {
	Liked bool `json:"liked"`
}

// FavoriteResult 收藏/取消收藏结果。
type FavoriteResult struct {
	Favorited bool `json:"favorited"`
}

// RatingResult 评分结果（重算后的主题平均分）。
type RatingResult struct {
	Rating float64 `json:"rating"`
}

// DefaultTheme 官方默认主题（GET /themes/default 返回，含制品下载地址）。
type DefaultTheme struct {
	ThemeItem
	IsDefault   bool   `json:"is_default"`
	DownloadURL string `json:"download_url"`
}

// Login 官方账号登录，返回 Token 与用户信息。
func Login(ctx context.Context, baseURL, email, password string) (TokenPair, UserInfo, error) {
	var data loginResp
	body := map[string]string{"email": email, "password": password}
	if err := call(ctx, baseURL, http.MethodPost, "/auth/login", "", body, &data); err != nil {
		return TokenPair{}, UserInfo{}, err
	}
	return data.Token, data.User, nil
}

// Logout 登出官方账号并拉黑当前 Token。
func Logout(ctx context.Context, baseURL, token string) error {
	return call(ctx, baseURL, http.MethodPost, "/auth/logout", token, nil, nil)
}

// ListThemes 查询官方主题列表（仅已上架主题）。
func ListThemes(ctx context.Context, baseURL, token string, q ThemeQuery) (ThemeList, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(q.Page))
	params.Set("size", strconv.Itoa(q.Size))
	if q.Type != "" {
		params.Set("type", q.Type)
	}
	for _, s := range q.Styles {
		params.Add("styles", s)
	}
	for _, f := range q.Features {
		params.Add("features", f)
	}
	if q.Price != "" {
		params.Set("price", q.Price)
	}
	if q.Search != "" {
		params.Set("search", q.Search)
	}
	if q.Sort != "" {
		params.Set("sort", q.Sort)
	}

	var data ThemeList
	path := "/themes?" + params.Encode()
	if err := call(ctx, baseURL, http.MethodGet, path, token, nil, &data); err != nil {
		return ThemeList{}, err
	}
	return data, nil
}

// GetTheme 查询主题详情（携带 token 时额外返回个人点赞/评分状态）。
func GetTheme(ctx context.Context, baseURL, id, token string) (ThemeDetail, error) {
	var data ThemeDetail
	if err := call(ctx, baseURL, http.MethodGet, "/themes/"+id, token, nil, &data); err != nil {
		return ThemeDetail{}, err
	}
	return data, nil
}

// GetThemeReleases 查询主题版本历史与更新日志（按发布时间倒序，无记录返回空切片）。
func GetThemeReleases(ctx context.Context, baseURL, id, token string) ([]ThemeReleaseItem, error) {
	var data []ThemeReleaseItem
	if err := call(ctx, baseURL, http.MethodGet, "/themes/"+id+"/releases", token, nil, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetStats 查询市场统计（主题总数/作者数/总安装次数）。
func GetStats(ctx context.Context, baseURL string) (ThemeStats, error) {
	var data ThemeStats
	if err := call(ctx, baseURL, http.MethodGet, "/themes/stats", "", nil, &data); err != nil {
		return ThemeStats{}, err
	}
	return data, nil
}

// GetHotTags 查询按上架主题风格聚合的热门标签（前 15 个）。
func GetHotTags(ctx context.Context, baseURL string) ([]HotTag, error) {
	var data []HotTag
	if err := call(ctx, baseURL, http.MethodGet, "/themes/tags", "", nil, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// LikeTheme 点赞/取消点赞（toggle），返回最新点赞状态。
func LikeTheme(ctx context.Context, baseURL, id, token string) (LikeResult, error) {
	var data LikeResult
	if err := call(ctx, baseURL, http.MethodPost, "/themes/"+id+"/like", token, nil, &data); err != nil {
		return LikeResult{}, err
	}
	return data, nil
}

// FavoriteTheme 收藏/取消收藏（toggle），返回最新收藏状态。
func FavoriteTheme(ctx context.Context, baseURL, id, token string) (FavoriteResult, error) {
	var data FavoriteResult
	if err := call(ctx, baseURL, http.MethodPost, "/themes/"+id+"/favorite", token, nil, &data); err != nil {
		return FavoriteResult{}, err
	}
	return data, nil
}

// RateTheme 为主题评分（1-5 分，同一用户重复评分覆盖旧分），返回重算后的平均分。
func RateTheme(ctx context.Context, baseURL, id string, score int, token string) (RatingResult, error) {
	var data RatingResult
	body := map[string]int{"score": score}
	if err := call(ctx, baseURL, http.MethodPost, "/themes/"+id+"/rating", token, body, &data); err != nil {
		return RatingResult{}, err
	}
	return data, nil
}

// IsOfficialDownloadURL 判断 rawURL 是否为官方代理下载地址（{官方API}/api/v1/themes/{id}/download）。
// 仅当 host 与官方 base 完全一致时才认定为官方地址：调用方据此决定是否携带官方 Token，
// 避免把凭据发送到第三方 host。其他形态（.tar.gz 直链、GitHub 仓库目录地址等）返回 false。
func IsOfficialDownloadURL(baseURL, rawURL string) bool {
	if baseURL == "" || rawURL == "" {
		return false
	}
	base, err := NormalizeBaseURL(baseURL)
	if err != nil {
		return false
	}
	baseU, err := url.Parse(base)
	if err != nil {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, baseU.Host) && officialDownloadPathPattern.MatchString(u.Path)
}

// ResolveDownload 请求官方代理下载地址（GET /themes/{id}/download）：
// 官方计数 +1 后 302 重定向到真实下载地址，本方法不跟随重定向，解析并返回 Location。
// token 必填（官方下载代理要求登录态）；401 → AuthError，404 等 → APIError。
func ResolveDownload(ctx context.Context, rawURL, token string) (string, error) {
	if rawURL == "" {
		return "", errors.New("官方下载地址不能为空")
	}
	if token == "" {
		return "", &AuthError{Message: "官方下载代理需要登录态"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("构建下载代理请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := downloadResolveClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求官方下载代理失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("官方下载代理未返回重定向地址")
		}
		// Location 理论上为绝对地址，防御性解析相对地址
		if ref, perr := req.URL.Parse(location); perr == nil {
			location = ref.String()
		}
		return location, nil
	}

	// 非 302：响应体为官方统一 JSON 错误结构（也可能为空），尽力解析
	var envelope officialResp
	_ = json.NewDecoder(resp.Body).Decode(&envelope)

	switch {
	case resp.StatusCode == http.StatusUnauthorized || envelope.Code == http.StatusUnauthorized:
		msg := envelope.Message
		if msg == "" {
			msg = "官方账号登录已失效"
		}
		return "", &AuthError{Message: msg}
	case resp.StatusCode == http.StatusOK || envelope.Code == http.StatusOK:
		return "", fmt.Errorf("官方下载代理未按预期重定向(HTTP %d)", resp.StatusCode)
	default:
		msg := envelope.Message
		if msg == "" {
			msg = resp.Status
		}
		return "", &APIError{Code: envelope.Code, Message: msg}
	}
}

// ListFavorites 查询当前用户收藏的主题列表（按收藏时间倒序，已删除主题自动跳过）。
func ListFavorites(ctx context.Context, baseURL, token string, page, size int) (ThemeList, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(size))

	var data ThemeList
	path := "/themes/favorites?" + params.Encode()
	if err := call(ctx, baseURL, http.MethodGet, path, token, nil, &data); err != nil {
		return ThemeList{}, err
	}
	return data, nil
}

// ListFavoriteIDs 查询当前用户收藏的全部主题 ID（用于批量点亮收藏状态）。
func ListFavoriteIDs(ctx context.Context, baseURL, token string) ([]int64, error) {
	var data []int64
	if err := call(ctx, baseURL, http.MethodGet, "/themes/favorites/ids", token, nil, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetDefaultTheme 查询官方默认主题（部署首装直接拉取；未设置或已下架时官方返回 404）。
func GetDefaultTheme(ctx context.Context, baseURL string) (DefaultTheme, error) {
	var data DefaultTheme
	if err := call(ctx, baseURL, http.MethodGet, "/themes/default", "", nil, &data); err != nil {
		return DefaultTheme{}, err
	}
	return data, nil
}
