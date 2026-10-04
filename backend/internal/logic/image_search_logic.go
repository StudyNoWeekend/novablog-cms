package logic

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"novablog/internal/dto/res"
	"novablog/pkg/bangumi"
	"novablog/pkg/bingimages"
	"novablog/pkg/doubanbooks"
	"novablog/pkg/proxyhttp"
	"novablog/pkg/simpleicons"
)

// ImageSearchSettings 图片搜索模块配置，由 bootstrap 阶段注入。
type ImageSearchSettings struct {
	ProxyURL string // 可选 HTTP 代理（空=直连；代理不可用自动降级直连）
}

// imageSearchSettings 图片搜索模块运行时配置。
var imageSearchSettings = &ImageSearchSettings{}

// SetImageSearchSettings 注入图片搜索模块配置。
func SetImageSearchSettings(s *ImageSearchSettings) {
	if s != nil {
		imageSearchSettings = s
	}
}

// ImageSearchLogic 图片搜索业务逻辑结构体。
type ImageSearchLogic struct {
	mediaLogic   *MediaLogic         // 转存复用媒体存储链路（含本地存储降级）
	doubanClient *doubanbooks.Client // 豆瓣客户端（CookieJar 跨请求保留 bid，过图片反爬）
}

// NewImageSearchLogic 创建 ImageSearchLogic 实例。
func NewImageSearchLogic(mediaLogic *MediaLogic) *ImageSearchLogic {
	return &ImageSearchLogic{
		mediaLogic:   mediaLogic,
		doubanClient: &doubanbooks.Client{},
	}
}

// douban 返回豆瓣客户端。豆瓣是国内站点，永远直连：
// 经代理的出口 IP 反而会概率性触发其反爬（返回 200 + HTML 验证页）。
func (l *ImageSearchLogic) douban() *doubanbooks.Client {
	l.doubanClient.ProxyURL = ""
	return l.doubanClient
}

// SearchImages 统一图片搜索：icons=品牌图标(Simple Icons)、games=游戏封面(Bangumi)、
// books=书籍封面(豆瓣读书)、food=美食图片(Bing 图片)。
func (l *ImageSearchLogic) SearchImages(ctx context.Context, typ, keyword string, limit int) ([]res.ImageSearchResult, error) {
	switch typ {
	case "icons":
		icons, err := (&simpleicons.Client{ProxyURL: imageSearchSettings.ProxyURL}).Search(ctx, keyword, limit)
		if err != nil {
			return nil, err
		}
		out := make([]res.ImageSearchResult, 0, len(icons))
		for _, i := range icons {
			out = append(out, res.ImageSearchResult{Name: i.Title, URL: i.URL})
		}
		return out, nil
	case "games":
		return l.searchSubjectCovers(ctx, keyword, 4, limit)
	case "books":
		covers, err := l.douban().Search(ctx, keyword, limit)
		if err != nil {
			return nil, err
		}
		out := make([]res.ImageSearchResult, 0, len(covers))
		for _, c := range covers {
			out = append(out, res.ImageSearchResult{Name: c.Name, URL: c.URL})
		}
		return out, nil
	case "food":
		// cn.bing.com 与大部分图片来源是国内 CDN，直连即可
		images, err := (&bingimages.Client{}).Search(ctx, keyword, limit)
		if err != nil {
			return nil, err
		}
		out := make([]res.ImageSearchResult, 0, len(images))
		for _, c := range images {
			out = append(out, res.ImageSearchResult{Name: c.Name, URL: c.URL})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("不支持的图片搜索类型: %s", typ)
	}
}

// searchSubjectCovers 按 Bangumi 条目类型搜索封面（4=游戏）。
func (l *ImageSearchLogic) searchSubjectCovers(ctx context.Context, keyword string, subjectType, limit int) ([]res.ImageSearchResult, error) {
	covers, err := (&bangumi.Client{ProxyURL: imageSearchSettings.ProxyURL}).SearchSubjects(ctx, keyword, subjectType, limit)
	if err != nil {
		return nil, err
	}
	out := make([]res.ImageSearchResult, 0, len(covers))
	for _, c := range covers {
		out = append(out, res.ImageSearchResult{Name: c.Name, URL: c.URL})
	}
	return out, nil
}

// SaveImage 将外部搜索到的图片转存到当前存储。豆瓣图床有 Cookie 反爬、Bing 图片源自
// 第三方站点存在防盗链，直接外链在博客前台可能失效，选中后统一转存为本站资源。
// module 为归属业务模块 key（如 game/book/recipe/tech_stack），转存后归入对应模块文件夹。
func (l *ImageSearchLogic) SaveImage(ctx context.Context, remoteURL, module string) (*res.MediaRes, error) {
	data, contentType, err := l.FetchImageBytes(ctx, remoteURL)
	if err != nil {
		return nil, err
	}

	ext := imageExt(contentType, remoteURL)
	if ext == "" {
		return nil, fmt.Errorf("不支持的图片格式（Content-Type: %s）", contentType)
	}
	filename := fmt.Sprintf("search-cover-%s%s", time.Now().Format("20060102150405"), ext)
	return l.mediaLogic.SaveRemoteBytes(ctx, filename, getMimeType(ext), data, MediaUploadOptions{Module: module})
}

// FetchImageBytes 抓取外部图片字节，按图源选择网络路径：
// 豆瓣（Cookie 反爬，恒直连）→ Bangumi 图床（部分网络直连不可达，走代理）→ 其余直连。
func (l *ImageSearchLogic) FetchImageBytes(ctx context.Context, remoteURL string) ([]byte, string, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return nil, "", fmt.Errorf("图片地址不合法")
	}

	switch {
	case strings.Contains(remoteURL, "doubanio.com"):
		return l.douban().DownloadCover(ctx, remoteURL)
	default:
		// lain.bgm.tv 等部分图床在国内网络直连不可达，需经代理；其余来源直连
		proxy := ""
		if strings.Contains(remoteURL, "bgm.tv") {
			proxy = imageSearchSettings.ProxyURL
		}
		return fetchRemoteImage(ctx, remoteURL, proxy)
	}
}

// Thumbnail 代理输出外部图片缩略图字节，供前端搜索弹窗 <img> 直接引用，
// 避免浏览器直连外部图床触发防盗链/反爬（豆瓣 418）或直连不可达（lain.bgm.tv）。
// 安全防护：仅 http/https，且解析后拒绝环回/内网/链路本地地址（防 SSRF）。
func (l *ImageSearchLogic) Thumbnail(ctx context.Context, rawURL string) ([]byte, string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, "", fmt.Errorf("图片地址不合法")
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return nil, "", fmt.Errorf("图片主机解析失败: %w", err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
			return nil, "", fmt.Errorf("不允许访问内网地址")
		}
	}
	return l.FetchImageBytes(ctx, rawURL)
}

// fetchRemoteImage 通用远程图片抓取（带浏览器 UA，适配部分站点的防盗链；proxyURL 为空时直连）。
func fetchRemoteImage(ctx context.Context, remoteURL, proxyURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")

	client := proxyhttp.NewResilient(proxyhttp.Config{ProxyURL: proxyURL, Timeout: 20 * time.Second})
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("图片下载返回状态码 %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[:idx]
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("链接不是图片（Content-Type: %s）", contentType)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}

// imageExt 从 Content-Type 或 URL 路径推断图片扩展名，仅接受媒体模块支持的类型。
func imageExt(contentType, remoteURL string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	}
	u := strings.ToLower(strings.SplitN(remoteURL, "?", 2)[0])
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp"} {
		if strings.HasSuffix(u, ext) {
			if ext == ".jpeg" {
				return ".jpg"
			}
			return ext
		}
	}
	return ""
}
