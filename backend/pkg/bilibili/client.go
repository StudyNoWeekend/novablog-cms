// Package bilibili 提供 B 站视频信息与播放地址解析能力。
//
// 所有公开方法均接受 context.Context 并以 error 形式返回失败原因。
// 业务缓存策略（Redis/TTL）由 logic 层编排，本包不直接访问任何缓存。
package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// client 公用 HTTP 客户端，统一设置超时。
var client = &http.Client{Timeout: 15 * time.Second}

// userAgent 浏览器 UA，避免被 B 站风控拦截。
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// referer 浏览器来源，部分接口要求 Referer 为 www.bilibili.com。
const referer = "https://www.bilibili.com"

// bvidRegex 校验 BV 号格式：BV 前缀 + 10 位字母数字。
var bvidRegex = regexp.MustCompile(`BV[a-zA-Z0-9]{10}`)

// VideoPage 视频分P信息。
type VideoPage struct {
	CID      int64  // 分P 的 cid
	Part     string // 分P 标题
	Duration int    // 分P 时长（秒）
}

// VideoInfo B站视频元信息。
type VideoInfo struct {
	Title     string      // 视频主标题（多分P 时为第一P）
	Pic       string      // 封面图
	Desc      string      // 简介
	OwnerName string      // UP 主名
	CID       int64       // 默认分P 的 cid
	Duration  int         // 默认分P 时长
	Pages     []VideoPage // 全部分P
}

// BuildEmbedURL 构造 B 站官方外链播放器地址，供 iframe 内嵌播放。
// 官方外链无需解析 CDN 直链，不受 Referer 防盗链与链接时效限制。
// autoplay=1 配合 iframe allow="autoplay" 在用户点击歌曲后自动播放；
// danmaku=0 关闭弹幕；high_quality=1 请求较高清晰度。
func BuildEmbedURL(bvid string) string {
	return fmt.Sprintf("https://player.bilibili.com/player.html?bvid=%s&autoplay=1&danmaku=0&high_quality=1", bvid)
}

// newRequest 创建带 UA/Referer 的 GET 请求。
func newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)
	return req, nil
}

// ExtractBVID 从多种 B 站 URL 格式中提取 BV 号。
// 不支持 b23.tv 短链（无法在服务端直接展开）。
func ExtractBVID(url string) (string, error) {
	if matched, _ := regexp.MatchString(`^https?://b23\.tv/`, url); matched {
		return "", fmt.Errorf("暂不支持 b23.tv 短链解析，请提供完整视频地址")
	}
	// 从 URL 中查找 BV 号（兼容 query/hash 中包含的场景）。
	bvid := bvidRegex.FindString(url)
	if bvid == "" {
		return "", fmt.Errorf("无法从 URL 中提取 BV 号: %s", url)
	}
	return bvid, nil
}

// biliVideoInfoResp /x/web-interface/view 响应。
type biliVideoInfoResp struct {
	Code int `json:"code"`
	Data struct {
		Title    string `json:"title"`
		Pic      string `json:"pic"`
		Desc     string `json:"desc"`
		CID      int64  `json:"cid"`
		Duration int    `json:"duration"`
		Owner    struct {
			Name string `json:"name"`
		} `json:"owner"`
		Pages []struct {
			CID      int64  `json:"cid"`
			Part     string `json:"part"`
			Duration int    `json:"duration"`
		} `json:"pages"`
	} `json:"data"`
}

// upgradePicURL 将 B 站返回的 http 封面 URL 升级为 https，
// 避免在 HTTPS 站点中出现混合内容（Mixed Content）被浏览器拦截。
func upgradePicURL(pic string) string {
	if strings.HasPrefix(pic, "http://") {
		return "https://" + pic[len("http://"):]
	}
	return pic
}

// FetchVideoInfo 调用 B 站 /x/web-interface/view 拉取视频元信息。
// 官方文档: https://socialsisteryi.github.io/bilibili-API-collect/docs/video/info.html
func FetchVideoInfo(ctx context.Context, bvid string) (*VideoInfo, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	req, err := newRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求视频信息失败: %w", err)
	}
	defer resp.Body.Close()

	var body biliVideoInfoResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("解析视频信息响应失败: %w", err)
	}
	if body.Code != 0 {
		return nil, fmt.Errorf("获取视频信息失败, code: %d", body.Code)
	}

	info := &VideoInfo{
		Title:     body.Data.Title,
		Pic:       upgradePicURL(body.Data.Pic),
		Desc:      body.Data.Desc,
		OwnerName: body.Data.Owner.Name,
		CID:       body.Data.CID,
		Duration:  body.Data.Duration,
	}
	for _, p := range body.Data.Pages {
		info.Pages = append(info.Pages, VideoPage{
			CID:      p.CID,
			Part:     p.Part,
			Duration: p.Duration,
		})
	}
	return info, nil
}
