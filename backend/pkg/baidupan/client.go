// Package baidupan 实现百度网盘开放平台（xpan OpenAPI）客户端。
// 文档：https://pan.baidu.com/union/doc/。access_token 有效期 30 天，refresh_token 10 年；
// 下载直链 dlink 有效期 8 小时，请求时必须携带以 pan.baidu.com 开头的 User-Agent。
package baidupan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"novablog/pkg/clouddrive"
	"novablog/pkg/proxyhttp"
)

// 开放平台端点。
const (
	APIBase     = "https://pan.baidu.com/rest/2.0"
	OAuthBase   = "https://openapi.baidu.com/oauth/2.0"
	AuthorizeURL = OAuthBase + "/authorize"
	TokenURL    = OAuthBase + "/token"

	// downloadUA 下载 dlink 时必须携带的 User-Agent（开放平台要求 pan.baidu.com 开头）。
	downloadUA = "pan.baidu.com;novaCMS;1.0.0"
)

// Client 百度网盘客户端。
type Client struct {
	AccessToken string
	ProxyURL    string
}

// New 创建百度网盘客户端。
func New(accessToken, proxyURL string) *Client {
	return &Client{AccessToken: accessToken, ProxyURL: proxyURL}
}

// baiduResp 列表/文件信息接口的通用响应（HTTP 200 + errno 错误码）。
type baiduResp struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

// apiClient API 调用客户端（30s 超时）。
func (c *Client) apiClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: 30 * time.Second})
}

// downloadClient 下载客户端：无整体超时，流式传输由 ctx 控制取消。
func (c *Client) downloadClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL})
}

// listEntry 列表接口的单个文件条目。
type listEntry struct {
	FsID           int64  `json:"fs_id"`
	Path           string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Category       int    `json:"category"`
	IsDir          int    `json:"isdir"`
	Size           int64  `json:"size"`
	MD5            string `json:"md5"`
}

type listResp struct {
	baiduResp
	List []listEntry `json:"list"`
}

// List 列出目录内容。parent 为盘内路径（根目录传 "/"），cursor 为页码（首页传空）。
func (c *Client) List(ctx context.Context, parent string, cursor string, limit int) ([]clouddrive.File, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	page := 1
	if cursor != "" {
		if p, err := strconv.Atoi(cursor); err == nil && p > 0 {
			page = p
		}
	}
	if parent == "" {
		parent = "/"
	}

	params := url.Values{}
	params.Set("method", "list")
	params.Set("access_token", c.AccessToken)
	params.Set("dir", parent)
	params.Set("web", "1")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("page", strconv.Itoa(page))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIBase+"/xpan/file?"+params.Encode(), nil)
	if err != nil {
		return nil, "", err
	}
	body, err := doJSON(c.apiClient(), req)
	if err != nil {
		return nil, "", err
	}
	var resp listResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("解析百度网盘响应失败: %w", err)
	}
	if resp.Errno != 0 {
		return nil, "", baiduErr(resp.Errno)
	}

	files := make([]clouddrive.File, 0, len(resp.List))
	for _, e := range resp.List {
		name := e.ServerFilename
		if name == "" {
			name = pathBase(e.Path)
		}
		f := clouddrive.File{
			FileID:   strconv.FormatInt(e.FsID, 10),
			ParentID: parent,
			Path:     e.Path,
			Name:     name,
			IsDir:    e.IsDir == 1,
			Size:     e.Size,
		}
		if !f.IsDir {
			f.Ext = strings.ToLower(pathExt(name))
		}
		files = append(files, f)
	}
	next := ""
	if len(resp.List) >= limit {
		next = strconv.Itoa(page + 1)
	}
	return files, next, nil
}

// metasResp filemetas 接口响应（含下载直链）。
type metasResp struct {
	baiduResp
	List []metaEntry `json:"list"`
}

type metaEntry struct {
	FsID     int64  `json:"fs_id"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Dlink    string `json:"dlink"`
}

// GetDownloadURL 获取文件下载直链（8 小时有效，调用方应即取即用）。
func (c *Client) GetDownloadURL(ctx context.Context, fileID string) (string, int64, error) {
	params := url.Values{}
	params.Set("method", "filemetas")
	params.Set("access_token", c.AccessToken)
	params.Set("fsids", "["+fileID+"]")
	params.Set("dlink", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIBase+"/xpan/file?"+params.Encode(), nil)
	if err != nil {
		return "", 0, err
	}
	body, err := doJSON(c.apiClient(), req)
	if err != nil {
		return "", 0, err
	}
	var resp metasResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", 0, fmt.Errorf("解析百度网盘响应失败: %w", err)
	}
	if resp.Errno != 0 {
		return "", 0, baiduErr(resp.Errno)
	}
	if len(resp.List) == 0 || resp.List[0].Dlink == "" {
		return "", 0, fmt.Errorf("未获取到下载直链")
	}
	return resp.List[0].Dlink, resp.List[0].Size, nil
}

// Download 获取文件下载流：filemetas 取 dlink → 携带约定 UA 请求直链。
func (c *Client) Download(ctx context.Context, fileID string) (io.ReadCloser, int64, string, error) {
	dlink, size, err := c.GetDownloadURL(ctx, fileID)
	if err != nil {
		return nil, 0, "", err
	}
	sep := "&"
	if !strings.Contains(dlink, "?") {
		sep = "?"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dlink+sep+"access_token="+url.QueryEscape(c.AccessToken), nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("User-Agent", downloadUA)

	resp, err := c.downloadClient().Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("下载文件失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, "", fmt.Errorf("下载文件返回状态码 %d", resp.StatusCode)
	}
	if size <= 0 {
		size = resp.ContentLength
	}
	return resp.Body, size, resp.Header.Get("Content-Type"), nil
}

// uinfoResp 用户信息接口响应。
type uinfoResp struct {
	BaiduName    string `json:"baidu_name"`
	NetdiskName  string `json:"netdisk_name"`
	AvatarURL    string `json:"avatar_url"`
	NetdiskQuota int64  `json:"netdisk_quota"`
	NetdiskUsed  int64  `json:"netdisk_used"`
}

// UserInfo 获取账号信息（昵称与容量）。
func (c *Client) UserInfo(ctx context.Context) (clouddrive.UserInfo, error) {
	params := url.Values{}
	params.Set("method", "uinfo")
	params.Set("access_token", c.AccessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIBase+"/xpan/nas?"+params.Encode(), nil)
	if err != nil {
		return clouddrive.UserInfo{}, err
	}
	body, err := doJSON(c.apiClient(), req)
	if err != nil {
		return clouddrive.UserInfo{}, err
	}
	var resp uinfoResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return clouddrive.UserInfo{}, fmt.Errorf("解析百度网盘用户信息失败: %w", err)
	}
	display := resp.BaiduName
	if display == "" {
		display = resp.NetdiskName
	}
	return clouddrive.UserInfo{
		OpenUserID:  display,
		DisplayName: display,
		QuotaTotal:  resp.NetdiskQuota,
		QuotaUsed:   resp.NetdiskUsed,
	}, nil
}

// ExchangeToken 用授权码换取凭据（grant_type=authorization_code）。
func ExchangeToken(ctx context.Context, clientID, clientSecret, redirectURI, code, proxyURL string) (*clouddrive.TokenResult, error) {
	return tokenRequest(ctx, map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"redirect_uri":  redirectURI,
	}, proxyURL)
}

// RefreshToken 刷新凭据（grant_type=refresh_token）。
func RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken, proxyURL string) (*clouddrive.TokenResult, error) {
	return tokenRequest(ctx, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}, proxyURL)
}

// tokenRequest 调用开放平台 token 接口。
func tokenRequest(ctx context.Context, params map[string]string, proxyURL string) (*clouddrive.TokenResult, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TokenURL+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	body, err := doJSON(proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: proxyURL, Timeout: 30 * time.Second}), req)
	if err != nil {
		return nil, err
	}
	var resp struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析百度网盘凭据响应失败: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("百度网盘授权失败: %s（%s）", resp.Error, resp.ErrorDescription)
	}
	if resp.AccessToken == "" {
		return nil, fmt.Errorf("百度网盘授权失败：未返回 access_token")
	}
	return &clouddrive.TokenResult{AccessToken: resp.AccessToken, RefreshToken: resp.RefreshToken, ExpiresIn: resp.ExpiresIn}, nil
}

// BuildAuthorizeURL 构造授权页地址（scope 申请网盘读写基础权限，支持扫码）。
func BuildAuthorizeURL(clientID, redirectURI, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "basic,netdisk")
	params.Set("display", "page")
	if state != "" {
		params.Set("state", state)
	}
	return AuthorizeURL + "?" + params.Encode()
}

// doJSON 执行请求并读取响应体（百度接口 HTTP 200 + errno 的模式在调用方处理）。
func doJSON(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求百度网盘接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("百度网盘接口返回状态码 %d", resp.StatusCode)
	}
	return body, nil
}

// baiduErr 将 errno 转为可读错误。
func baiduErr(errno int) error {
	switch errno {
	case 111:
		return fmt.Errorf("access_token 已失效，请重新授权")
	case -6:
		return fmt.Errorf("百度账号身份校验失败，请重新授权")
	default:
		return fmt.Errorf("百度网盘接口错误（errno=%d）", errno)
	}
}

// pathBase / pathExt 简单的路径工具（避免引入 filepath 处理盘内路径）。
func pathBase(p string) string {
	p = strings.TrimRight(p, "/")
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		return p[idx+1:]
	}
	return p
}

func pathExt(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx:]
	}
	return ""
}
