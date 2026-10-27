// Package bilibili 提供 B 站视频信息与播放地址解析能力。
//
// 所有公开方法均接受 context.Context 并以 error 形式返回失败原因。
// 业务缓存策略（Redis/TTL）由 logic 层编排，本包不直接访问任何缓存。
package bilibili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// client 公用 HTTP 客户端，统一设置超时。
var client = &http.Client{Timeout: 15 * time.Second}

// apiBaseURL B 站 web 接口基地址，测试时可替换为 mock 服务地址。
var apiBaseURL = "https://api.bilibili.com"

// maxRespBodyBytes 响应体读取上限，防止异常响应占用过多内存。
const maxRespBodyBytes = 8 << 20

// userAgent 浏览器 UA，避免被 B 站风控拦截。
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// referer 浏览器来源，部分接口要求 Referer 为 www.bilibili.com。
const referer = "https://www.bilibili.com"

// errCodeWBIInvalid wbi 签名校验失败的错误码，刷新密钥重试即可恢复。
const errCodeWBIInvalid = -403

// bvidRegex 校验 BV 号格式：BV 前缀 + 10 位字母数字。
var bvidRegex = regexp.MustCompile(`BV[a-zA-Z0-9]{10}`)

// biliAPIError 接口可达但业务码非零（如视频不存在），属接口明确回答，降级重试无意义。
type biliAPIError struct {
	code int
}

// Error 实现 error 接口。
func (e *biliAPIError) Error() string {
	return fmt.Sprintf("获取视频信息失败, code: %d", e.code)
}

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

// biliVideoInfoResp wbi/view 响应，结构与旧版 view 接口一致。
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

// biliPagelistResp /x/player/pagelist 响应，降级解析用。
type biliPagelistResp struct {
	Code int `json:"code"`
	Data []struct {
		CID      int64  `json:"cid"`
		Part     string `json:"part"`
		Duration int    `json:"duration"`
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

// FetchVideoInfo 获取视频元信息。
// 首选官方 WBI 签名接口 /x/web-interface/wbi/view（完整元信息）；被 412 风控等
// 拦截时降级 /x/player/pagelist（仅分P信息：标题取第一P的分P标题，无封面/简介/UP主）。
// 两级都失败时返回包含各自原因的错误。
func FetchVideoInfo(ctx context.Context, bvid string) (*VideoInfo, error) {
	info, err := fetchVideoInfoWBI(ctx, bvid)
	if err == nil {
		return info, nil
	}
	var apiErr *biliAPIError
	if errors.As(err, &apiErr) {
		// 接口明确回答的业务错误（如视频不存在），与风控无关，无需降级
		return nil, err
	}
	info, pagelistErr := fetchVideoInfoPagelist(ctx, bvid)
	if pagelistErr == nil {
		return info, nil
	}
	return nil, fmt.Errorf("B站视频解析失败: %v（降级 pagelist 也失败: %v）", err, pagelistErr)
}

// fetchVideoInfoWBI 调 WBI 签名的 view 接口；签名失效（-403）时刷新密钥重试一次。
func fetchVideoInfoWBI(ctx context.Context, bvid string) (*VideoInfo, error) {
	info, err := fetchVideoInfoWBIOnce(ctx, bvid)
	if err == nil {
		return info, nil
	}
	var apiErr *biliAPIError
	if !errors.As(err, &apiErr) || apiErr.code != errCodeWBIInvalid {
		return nil, err
	}
	if _, refreshErr := getMixinKey(ctx, true); refreshErr != nil {
		return nil, fmt.Errorf("刷新 WBI 签名密钥失败: %w", refreshErr)
	}
	return fetchVideoInfoWBIOnce(ctx, bvid)
}

// fetchVideoInfoWBIOnce 单次调用 WBI 签名的 view 接口。
func fetchVideoInfoWBIOnce(ctx context.Context, bvid string) (*VideoInfo, error) {
	mixinKey, err := getMixinKey(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("获取 WBI 签名密钥失败: %w", err)
	}
	signed := signWBI(map[string]string{"bvid": bvid}, mixinKey)
	body, err := doGet(ctx, apiBaseURL+"/x/web-interface/wbi/view?"+signed, getBuvidCookie(ctx))
	if err != nil {
		return nil, err
	}
	var resp biliVideoInfoResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析视频信息响应失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, &biliAPIError{code: resp.Code}
	}
	return videoInfoFromView(&resp), nil
}

// fetchVideoInfoPagelist 降级解析：调 /x/player/pagelist 取分P信息组装元信息。
func fetchVideoInfoPagelist(ctx context.Context, bvid string) (*VideoInfo, error) {
	body, err := doGet(ctx, apiBaseURL+"/x/player/pagelist?bvid="+url.QueryEscape(bvid), "")
	if err != nil {
		return nil, err
	}
	var resp biliPagelistResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析分P信息响应失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("获取分P信息失败, code: %d", resp.Code)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("视频无可用分P信息")
	}
	info := &VideoInfo{Title: resp.Data[0].Part, CID: resp.Data[0].CID}
	for _, p := range resp.Data {
		info.Duration += p.Duration
		info.Pages = append(info.Pages, VideoPage{CID: p.CID, Part: p.Part, Duration: p.Duration})
	}
	return info, nil
}

// videoInfoFromView 将 view 响应映射为 VideoInfo。
func videoInfoFromView(resp *biliVideoInfoResp) *VideoInfo {
	info := &VideoInfo{
		Title:     resp.Data.Title,
		Pic:       upgradePicURL(resp.Data.Pic),
		Desc:      resp.Data.Desc,
		OwnerName: resp.Data.Owner.Name,
		CID:       resp.Data.CID,
		Duration:  resp.Data.Duration,
	}
	for _, p := range resp.Data.Pages {
		info.Pages = append(info.Pages, VideoPage{CID: p.CID, Part: p.Part, Duration: p.Duration})
	}
	return info
}

// doGet 发起 GET 请求并返回响应体。412 风控与非 200 状态均返回带明确语义的错误，
// 避免把 HTML 验证页当 JSON 解析。
func doGet(ctx context.Context, reqURL, cookie string) ([]byte, error) {
	req, err := newRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusPreconditionFailed {
		return nil, fmt.Errorf("B站风控拦截（HTTP 412），请稍后重试或为服务器配置可用代理")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("响应状态异常, HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxRespBodyBytes))
}
