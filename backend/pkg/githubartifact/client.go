// Package githubartifact 解析 GitHub 主题仓库目录对应的 Release 预构建制品。
//
// 官方市场的 download_url 指向源码目录（形如 https://github.com/{owner}/{repo}/tree/HEAD/{dir}），
// 可托管的静态制品由主题作者通过 CI 发布到该仓库的 GitHub Release
// （资产命名约定：{dir}-{version}.tar.gz，另附 .sha256 校验文件）。
package githubartifact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"novablog/pkg/proxyhttp"
)

// apiBase GitHub REST API 地址（固定域，防 SSRF）。
const apiBase = "https://api.github.com"

// ErrArtifactNotFound Release 中未找到匹配的预构建制品。
var ErrArtifactNotFound = errors.New("release 中未找到匹配的预构建制品")

// ErrInvalidRepoURL download_url 不是合法的 GitHub 仓库目录地址。
var ErrInvalidRepoURL = errors.New("下载地址不是合法的 GitHub 仓库目录地址")

// repoDirPattern 匹配 https://github.com/{owner}/{repo}/tree/{ref}/{dir}（兼容 blob 与多级 ref）。
var repoDirPattern = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/(?:tree|blob)/(.+)$`)

// githubProbeURL 代理可用性探测端点（与 apiBase 同域，代表 GitHub 可达性）。
const githubProbeURL = "https://api.github.com/zen"

// httpClient 按当前配置构建 GitHub API 客户端：
// 配置了代理时走代理（探测缓存 + 代理中途失败自动降级直连），未配置则纯直连。
func (c *Client) httpClient() *http.Client {
	return proxyhttp.NewResilient(proxyhttp.Config{
		Timeout:  30 * time.Second,
		ProxyURL: c.ProxyURL,
		ProbeURL: githubProbeURL,
	})
}

// Asset Release 制品资产。
type Asset struct {
	Name        string // 资产文件名，如 tech-geek-0.1.0.tar.gz
	DownloadURL string // 浏览器下载地址
	Tag         string // 所属 Release 标签
	Version     string // 从资产名解析出的主题版本
}

// Client GitHub 制品解析客户端。
type Client struct {
	Token    string // 可选 PAT，提升限流额度
	ProxyURL string // 可选 HTTP 代理（空=直连；代理不可用自动降级直连）
}

// ghAsset GitHub Release 资产（API 响应子集）。
type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ghRelease GitHub Release（API 响应子集）。
type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

// ParseRepoDir 从官方 download_url 解析 owner、repo 与主题目录名。
func ParseRepoDir(raw string) (owner, repo, dir string, err error) {
	m := repoDirPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return "", "", "", ErrInvalidRepoURL
	}
	// ref 段可能多级（tree/HEAD/dir、tree/main/sub/dir），目录名取最后一段
	rest := m[3]
	idx := strings.LastIndex(rest, "/")
	dirName := rest
	if idx >= 0 {
		dirName = rest[idx+1:]
	}
	if dirName == "" {
		return "", "", "", ErrInvalidRepoURL
	}
	return m[1], m[2], dirName, nil
}

// FindReleaseAsset 在仓库最近的 Release 中查找匹配制品。
// version 非空时精确匹配 {dir}-{version}.tar.gz；为空时取最近 Release 中 {dir}-*.tar.gz。
func (c *Client) FindReleaseAsset(ctx context.Context, owner, repo, dir, version string) (Asset, error) {
	releases, err := c.listReleases(ctx, owner, repo)
	if err != nil {
		return Asset{}, err
	}

	exact := fmt.Sprintf("%s-%s.tar.gz", dir, version)
	for _, rel := range releases {
		for _, a := range rel.Assets {
			if version != "" {
				if a.Name == exact {
					return Asset{Name: a.Name, DownloadURL: a.BrowserDownloadURL, Tag: rel.TagName, Version: version}, nil
				}
				continue
			}
			if v, ok := parseAssetVersion(a.Name, dir); ok {
				return Asset{Name: a.Name, DownloadURL: a.BrowserDownloadURL, Tag: rel.TagName, Version: v}, nil
			}
		}
	}
	return Asset{}, ErrArtifactNotFound
}

// listReleases 拉取仓库 Release 列表（最近 30 个）。
func (c *Client) listReleases(ctx context.Context, owner, repo string) ([]ghRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/repos/%s/%s/releases?per_page=30", apiBase, owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("构建 GitHub 请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub Release 失败: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrArtifactNotFound
	case resp.StatusCode == http.StatusForbidden:
		return nil, errors.New("GitHub API 访问受限（限流），请配置 themes.github_token 后重试")
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("请求 GitHub Release 失败(HTTP %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 GitHub 响应失败: %w", err)
	}

	var releases []ghRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("解析 GitHub 响应失败: %w", err)
	}
	return releases, nil
}

// assetNamePattern 匹配 {dir}-{version}.tar.gz 并提取版本。
var assetNamePattern = regexp.MustCompile(`^(.+)-(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)\.tar\.gz$`)

// parseAssetVersion 从资产名解析版本，仅接受 {dir}-x.y.z.tar.gz 形式。
func parseAssetVersion(name, dir string) (string, bool) {
	m := assetNamePattern.FindStringSubmatch(name)
	if m == nil || m[1] != dir {
		return "", false
	}
	return m[2], true
}
