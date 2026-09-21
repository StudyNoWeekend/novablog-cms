// 本文件提供仓库元数据与 README 拉取能力，供"开源作品"模块使用：
// 仓库地址解析 → 元数据（描述/语言/Star 等）→ README 原文（Markdown）。

package githubartifact

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// repoURLPattern 匹配 GitHub 仓库地址（兼容 .git 后缀、尾随斜杠与 tree/blob 等子路径）。
var repoURLPattern = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/#?]+?)(?:\.git)?(?:/.*)?(?:[?#].*)?$`)

// ErrInvalidOpenRepoURL 仓库地址不是合法的 GitHub 仓库地址。
var ErrInvalidOpenRepoURL = fmt.Errorf("%w: 仓库地址", ErrInvalidRepoURL)

// ParseRepoURL 从仓库链接解析 owner 与 repo。
// 兼容 https://github.com/owner/repo、尾随斜杠、.git 后缀及 /tree/main 等子路径。
func ParseRepoURL(raw string) (owner, repo string, err error) {
	m := repoURLPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return "", "", ErrInvalidOpenRepoURL
	}
	return m[1], m[2], nil
}

// RepoMeta 仓库公开元数据（API 响应子集）。
type RepoMeta struct {
	Description string   `json:"description"` // 仓库描述
	Homepage    string   `json:"homepage"`    // 主页/演示地址
	Language    string   `json:"language"`    // 主语言
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"`   // 主题标签
	HTMLURL     string   `json:"html_url"` // 仓库地址
}

// ghRepo GitHub 仓库 API 响应子集。
type ghRepo struct {
	Description string   `json:"description"`
	Homepage    string   `json:"homepage"`
	Language    string   `json:"language"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"`
	HTMLURL     string   `json:"html_url"`
}

// FetchRepoMeta 拉取仓库公开元数据。
func (c *Client) FetchRepoMeta(ctx context.Context, owner, repo string) (*RepoMeta, error) {
	var gh ghRepo
	if err := c.getJSON(ctx, fmt.Sprintf("/repos/%s/%s", owner, repo), &gh); err != nil {
		return nil, err
	}
	return &RepoMeta{
		Description: gh.Description,
		Homepage:    gh.Homepage,
		Language:    gh.Language,
		Stars:       gh.Stars,
		Topics:      gh.Topics,
		HTMLURL:     gh.HTMLURL,
	}, nil
}

// FetchReadme 拉取仓库默认分支的 README 原文（Markdown），自动适配 README 命名变体。
func (c *Client) FetchReadme(ctx context.Context, owner, repo string) (string, error) {
	req, err := c.newRequest(ctx, fmt.Sprintf("/repos/%s/%s/readme", owner, repo), "application/vnd.github.raw+json")
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 GitHub README 失败: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return "", fmt.Errorf("仓库不存在或没有 README 文件")
	case resp.StatusCode == http.StatusForbidden:
		return "", fmt.Errorf("GitHub API 访问受限（限流），请配置 themes.github_token 后重试")
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("请求 GitHub README 失败(HTTP %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取 README 响应失败: %w", err)
	}
	return string(body), nil
}

// getJSON 请求 GitHub API 并把 JSON 响应解析到 out。
func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	req, err := c.newRequest(ctx, path, "application/vnd.github+json")
	if err != nil {
		return err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("GitHub 资源不存在(%s)", path)
	case resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("GitHub API 访问受限（限流），请配置 themes.github_token 后重试")
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("请求 GitHub API 失败(HTTP %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("读取 GitHub 响应失败: %w", err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析 GitHub 响应失败: %w", err)
	}
	return nil
}

// newRequest 构建带公共头的 GitHub API 请求。
func (c *Client) newRequest(ctx context.Context, path, accept string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+path, nil)
	if err != nil {
		return nil, fmt.Errorf("构建 GitHub 请求失败: %w", err)
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return req, nil
}
