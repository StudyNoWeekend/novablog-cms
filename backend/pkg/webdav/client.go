// Package webdav 实现标准 WebDAV 客户端（PROPFIND 列目录 + GET 下载流式传输），
// 覆盖坚果云、InfiniCloud、123 云盘（会员）以及自建 Alist（可聚合百度/阿里/夸克等全部网盘）。
package webdav

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"novablog/pkg/clouddrive"
	"novablog/pkg/proxyhttp"
)

// Client WebDAV 客户端。
type Client struct {
	BaseURL  string // 服务根地址，如 https://dav.jianguoyun.com/dav/
	Username string
	Password string
	ProxyURL string
}

// New 创建 WebDAV 客户端。
func New(baseURL, username, password, proxyURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Username: username, Password: password, ProxyURL: proxyURL}
}

// propfind 解析 WebDAV multistatus 响应的最小字段集（命名空间 URI 即字面量 "DAV:"）。
type propfind struct {
	XMLName xml.Name    `xml:"DAV: multistatus"`
	Resps   []respProps `xml:"DAV: response"`
}

type respProps struct {
	Href     string     `xml:"DAV: href"`
	Propstat []propstat `xml:"DAV: propstat"`
}

type propstat struct {
	Prop propFields `xml:"DAV: prop"`
}

type propFields struct {
	DisplayName      string `xml:"DAV: displayname"`
	GetContentLength int64  `xml:"DAV: getcontentlength"`
	ResourceType     struct {
		Collection *struct{} `xml:"DAV: collection"`
	} `xml:"DAV: resourcetype"`
}

// client 构建带超时与代理降级的 HTTP 客户端（API 调用；下载用 downloadClient）。
func (c *Client) client() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL, Timeout: 60 * time.Second})
}

// downloadClient 构建下载专用客户端：不设整体超时（大文件流式传输由 ctx 控制取消）。
func (c *Client) downloadClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: c.ProxyURL})
}

// buildURL 拼接服务地址与文件路径（对路径段做转义）。
func (c *Client) buildURL(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return c.BaseURL + "/"
	}
	segments := strings.Split(p, "/")
	escaped := make([]string, 0, len(segments))
	for _, s := range segments {
		escaped = append(escaped, url.PathEscape(s))
	}
	return c.BaseURL + "/" + strings.Join(escaped, "/")
}

// basePath 提取 BaseURL 的路径部分（如 https://dav.jianguoyun.com/dav → /dav），
// 用于把 href（含 base path 的绝对路径）换算为相对于服务根的路径。
func (c *Client) basePath() string {
	if u, err := url.Parse(c.BaseURL); err == nil {
		return strings.TrimRight(u.Path, "/")
	}
	return ""
}

// List 列出目录内容（Depth:1，返回全部直接子项，无游标分页）。
// parent 为盘内目录路径，根目录传 "" 或 "/"。
func (c *Client) List(ctx context.Context, parent string, cursor string, limit int) ([]clouddrive.File, string, error) {
	parent = normalizeDir(parent)
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.buildURL(parent), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Depth", "1")
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("WebDAV 列目录失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, "", fmt.Errorf("WebDAV 目录不存在: %s", parent)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, "", fmt.Errorf("WebDAV 账号或密码错误")
	}
	if resp.StatusCode != http.StatusMultiStatus && resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("WebDAV 列目录返回状态码 %d", resp.StatusCode)
	}

	var result propfind
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("解析 WebDAV 响应失败: %w", err)
	}

	selfPath := normalizeDir(parent)
	files := make([]clouddrive.File, 0, len(result.Resps))
	for _, r := range result.Resps {
		entryPath := c.relativePath(r.Href)
		if entryPath == "" {
			continue
		}
		name, isDir, size := parseEntry(r, entryPath)
		if name == "" {
			continue
		}
		if entryPath == selfPath {
			continue // 排除目录自身
		}
		ext := ""
		if !isDir {
			ext = strings.ToLower(path.Ext(name))
		}
		files = append(files, clouddrive.File{
			FileID:   entryPath,
			ParentID: normalizeDir(path.Dir(entryPath)),
			Path:     entryPath,
			Name:     name,
			IsDir:    isDir,
			Size:     size,
			Ext:      ext,
		})
	}
	return files, "", nil
}

// Download 以流式 GET 下载文件（fileID 即盘内路径）。
func (c *Client) Download(ctx context.Context, fileID string) (io.ReadCloser, int64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(fileID), nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.downloadClient().Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("WebDAV 下载失败: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, 0, "", fmt.Errorf("WebDAV 账号或密码错误")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, "", fmt.Errorf("WebDAV 下载返回状态码 %d", resp.StatusCode)
	}
	return resp.Body, resp.ContentLength, resp.Header.Get("Content-Type"), nil
}

// UserInfo 校验连通性；WebDAV 无标准容量接口，容量返回 0。
func (c *Client) UserInfo(ctx context.Context) (clouddrive.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.BaseURL+"/", nil)
	if err != nil {
		return clouddrive.UserInfo{}, err
	}
	req.Header.Set("Depth", "0")
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.client().Do(req)
	if err != nil {
		return clouddrive.UserInfo{}, fmt.Errorf("WebDAV 连接失败: %w", err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return clouddrive.UserInfo{}, fmt.Errorf("WebDAV 账号或密码错误")
	case resp.StatusCode == http.StatusNotFound:
		return clouddrive.UserInfo{}, fmt.Errorf("WebDAV 根目录不存在，请检查服务器地址")
	case resp.StatusCode != http.StatusMultiStatus && resp.StatusCode != http.StatusOK:
		return clouddrive.UserInfo{}, fmt.Errorf("WebDAV 连接返回状态码 %d", resp.StatusCode)
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return clouddrive.UserInfo{OpenUserID: c.Username, DisplayName: c.Username}, nil
}

// Test 校验连通性，供账号保存前测试使用。
func (c *Client) Test(ctx context.Context) error {
	_, err := c.UserInfo(ctx)
	return err
}

// parseEntry 从 propstat 中提取条目名、是否目录与大小。
func parseEntry(r respProps, entryPath string) (name string, isDir bool, size int64) {
	for _, ps := range r.Propstat {
		isDir = ps.Prop.ResourceType.Collection != nil
		size = ps.Prop.GetContentLength
		name = strings.TrimSpace(ps.Prop.DisplayName)
		break
	}
	if name == "" || name == "/" || name == "." {
		name = path.Base(entryPath)
	}
	if isDir {
		return name, true, 0
	}
	return name, false, size
}

// relativePath 将 href 换算为相对于服务根的路径（剥离 BaseURL 的路径前缀并解码）。
// 例：BaseURL=https://dav.jianguoyun.com/dav，href=/dav/sub/a.jpg → /sub/a.jpg。
func (c *Client) relativePath(href string) string {
	if href == "" {
		return ""
	}
	if u, err := url.Parse(href); err == nil {
		href = u.Path
	}
	base := c.basePath()
	if base != "" && strings.HasPrefix(href, base) {
		href = strings.TrimPrefix(href, base)
	}
	if href == "" || href == "/" {
		return "/"
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return strings.TrimRight(href, "/")
}

// normalizeDir 统一目录路径格式：以 / 开头、不以 / 结尾（根目录为 /）。
func normalizeDir(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if p == "/" {
		return "/"
	}
	return strings.TrimRight(p, "/")
}
