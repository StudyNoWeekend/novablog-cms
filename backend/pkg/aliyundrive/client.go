// Package aliyundrive 实现阿里云盘开放平台（adrive OpenAPI）客户端。
// 文档：https://www.alipan.com/developer/。access_token 有效期约 2 小时，refresh_token 长期有效；
// 全部请求必须携带在开放平台报备过的自定义 User-Agent。
package aliyundrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"novablog/pkg/clouddrive"
	"novablog/pkg/proxyhttp"
)

// 开放平台端点。
const (
	APIBase      = "https://openapi.alipan.com/adrive/v1.0"
	AuthorizeURL = "https://openauth.alipan.com/oauth/authorize"
	TokenURL     = "https://openauth.alipan.com/oauth/access_token"

	// Scope 授权范围：基础用户信息 + 文件读取（列表/下载）。
	Scope = "user.base file:read:list file:read:download"

	// UserAgent 开放平台报备的自定义 UA（格式：应用名/版本;端点/类型）。
	UserAgent = "novaCMS/1.0.0;cms/openapi"
)

// Client 阿里云盘客户端。
type Client struct {
	AccessToken string
	DriveID     string // default_drive_id
	ProxyURL    string
}

// New 创建阿里云盘客户端。
func New(accessToken, driveID, proxyURL string) *Client {
	return &Client{AccessToken: accessToken, DriveID: driveID, ProxyURL: proxyURL}
}

// apiClient API 调用客户端（30s 超时）。
func (c *Client) apiClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: 30 * time.Second})
}

// downloadClient 下载客户端：无整体超时，流式传输由 ctx 控制取消。
func (c *Client) downloadClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL})
}

// postJSON 发起 POST JSON 请求并解析响应。
func (c *Client) postJSON(ctx context.Context, endpoint string, payload interface{}, out interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIBase+endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.apiClient().Do(req)
	if err != nil {
		return fmt.Errorf("请求阿里云盘接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return aliyunErr(resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析阿里云盘响应失败: %w", err)
	}
	return nil
}

// listEntry 列表接口的单个条目。
type listEntry struct {
	FileID        string `json:"file_id"`
	ParentFileID  string `json:"parent_file_id"`
	Name          string `json:"name"`
	Type          string `json:"type"` // folder/file
	Size          int64  `json:"size"`
	FileExtension string `json:"file_extension"`
}

type listResp struct {
	Items      []listEntry `json:"items"`
	NextMarker string      `json:"next_marker"`
}

// List 列出目录内容。parent 为目录 file_id（根目录传空或 "root"），cursor 为 marker 游标。
func (c *Client) List(ctx context.Context, parent string, cursor string, limit int) ([]clouddrive.File, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	if parent == "" {
		parent = "root"
	}
	payload := map[string]interface{}{
		"drive_id":        c.DriveID,
		"parent_file_id":  parent,
		"limit":           limit,
		"marker":          cursor,
		"order_by":        "name",
		"order_direction": "ASC",
	}
	var resp listResp
	if err := c.postJSON(ctx, "/openFile/list", payload, &resp); err != nil {
		return nil, "", err
	}

	files := make([]clouddrive.File, 0, len(resp.Items))
	for _, e := range resp.Items {
		isDir := e.Type == "folder"
		f := clouddrive.File{
			FileID:   e.FileID,
			ParentID: e.ParentFileID,
			Name:     e.Name,
			IsDir:    isDir,
			Size:     e.Size,
		}
		if !isDir {
			f.Ext = normalizeExt(e.FileExtension)
		}
		files = append(files, f)
	}
	next := ""
	if resp.NextMarker != "" && len(resp.Items) >= limit {
		next = resp.NextMarker
	}
	return files, next, nil
}

// getDownloadURLResp 获取下载直链响应。
type getDownloadURLResp struct {
	URL string `json:"url"`
}

// GetDownloadURL 获取文件下载直链（数小时有效，即取即用）。
func (c *Client) GetDownloadURL(ctx context.Context, fileID string) (string, error) {
	var resp getDownloadURLResp
	err := c.postJSON(ctx, "/openFile/getDownloadUrl", map[string]interface{}{
		"drive_id": c.DriveID,
		"file_id":  fileID,
	}, &resp)
	if err != nil {
		return "", err
	}
	if resp.URL == "" {
		return "", fmt.Errorf("未获取到下载直链")
	}
	return resp.URL, nil
}

// Download 获取文件下载流。
func (c *Client) Download(ctx context.Context, fileID string) (io.ReadCloser, int64, string, error) {
	downloadURL, err := c.GetDownloadURL(ctx, fileID)
	if err != nil {
		return nil, 0, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.downloadClient().Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("下载文件失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, "", fmt.Errorf("下载文件返回状态码 %d", resp.StatusCode)
	}
	return resp.Body, resp.ContentLength, resp.Header.Get("Content-Type"), nil
}

// UserInfo 获取账号信息（用户昵称与默认云盘 ID，容量接口缺失时返回 0）。
func (c *Client) UserInfo(ctx context.Context) (clouddrive.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIBase+"/users/me", nil)
	if err != nil {
		return clouddrive.UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.apiClient().Do(req)
	if err != nil {
		return clouddrive.UserInfo{}, fmt.Errorf("请求阿里云盘接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return clouddrive.UserInfo{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return clouddrive.UserInfo{}, aliyunErr(resp.StatusCode, body)
	}
	var me struct {
		UserID         string `json:"user_id"`
		Name           string `json:"name"`
		DefaultDriveID string `json:"default_drive_id"`
	}
	if err := json.Unmarshal(body, &me); err != nil {
		return clouddrive.UserInfo{}, fmt.Errorf("解析阿里云盘用户信息失败: %w", err)
	}
	return clouddrive.UserInfo{
		OpenUserID:  me.UserID,
		DisplayName: me.Name,
		DriveID:     me.DefaultDriveID,
	}, nil
}

// ExchangeToken 用授权码换取凭据。
func ExchangeToken(ctx context.Context, clientID, clientSecret, redirectURI, code, proxyURL string) (*clouddrive.TokenResult, error) {
	return tokenRequest(ctx, map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"client_id":    clientID,
		"client_secret": clientSecret,
		"redirect_uri": redirectURI,
	}, proxyURL)
}

// RefreshToken 刷新凭据（阿里云盘 refresh_token 一次性，刷新后旧值作废，必须回存新值）。
func RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken, proxyURL string) (*clouddrive.TokenResult, error) {
	return tokenRequest(ctx, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}, proxyURL)
}

// tokenRequest 调用 OAuth token 接口（form-urlencoded）。
func tokenRequest(ctx context.Context, params map[string]string, proxyURL string) (*clouddrive.TokenResult, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: proxyURL, Timeout: 30 * time.Second})
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求阿里云盘凭据接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, aliyunErr(resp.StatusCode, body)
	}
	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("解析阿里云盘凭据响应失败: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("阿里云盘授权失败：未返回 access_token")
	}
	return &clouddrive.TokenResult{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, ExpiresIn: token.ExpiresIn}, nil
}

// BuildAuthorizeURL 构造授权页地址。
func BuildAuthorizeURL(clientID, redirectURI, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", Scope)
	if state != "" {
		params.Set("state", state)
	}
	return AuthorizeURL + "?" + params.Encode()
}

// aliyunErr 将非 200 响应转为可读错误（401=token 失效）。
func aliyunErr(statusCode int, body []byte) error {
	if statusCode == http.StatusUnauthorized {
		return fmt.Errorf("access_token 已失效，请重新授权")
	}
	msg := strings.TrimSpace(string(body))
	if len(msg) > 200 {
		msg = msg[:200]
	}
	return fmt.Errorf("阿里云盘接口返回状态码 %d：%s", statusCode, msg)
}

// normalizeExt 将不带点的扩展名归一为 ".ext"。
func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
