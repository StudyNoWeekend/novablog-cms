// WBI 请求签名：B 站官方 Web 端对 wbi 系列接口（如 /x/web-interface/wbi/view）
// 启用的请求签名机制，请求须携带 wts（秒级时间戳）与 w_rid（排序参数 + 密钥的 MD5），
// 无签名请求会被 412 风控拦截并返回 HTML 验证页。
// 算法与混淆表来源：https://socialsisteryi.github.io/bilibili-API-collect/docs/misc/sign/wbi.html

package bilibili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// mixinKeyEncTab 官方混淆置换表：将 nav 接口返回的 img/sub key 按下标重排得到真实密钥。
var mixinKeyEncTab = [...]int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40, 61,
	26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36,
	20, 34, 44, 52,
}

const (
	// mixinKeyLen 签名密钥长度。
	mixinKeyLen = 32
	// wbiKeyTTL nav 密钥缓存时长，官方密钥不定期轮换，不宜永久缓存。
	wbiKeyTTL = time.Hour
	// buvidCookieTTL spi Cookie 缓存时长。
	buvidCookieTTL = 24 * time.Hour
)

var (
	wbiKeyMu       sync.Mutex
	wbiKeyCache    string
	wbiKeyExpireAt time.Time

	buvidMu          sync.Mutex
	buvidCookieCache string
	buvidExpireAt    time.Time
)

// navResp /x/web-interface/nav 响应，匿名调用即含 wbi_img。
type navResp struct {
	Data struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

// spiResp /x/frontend/finger/spi 响应，向访客发放 buvid 身份 Cookie。
type spiResp struct {
	Code int `json:"code"`
	Data struct {
		B3 string `json:"b_3"`
		B4 string `json:"b_4"`
	} `json:"data"`
}

// getMixinKey 返回 WBI 签名密钥，带进程内缓存；forceRefresh 用于签名失效后强制轮换。
func getMixinKey(ctx context.Context, forceRefresh bool) (string, error) {
	wbiKeyMu.Lock()
	defer wbiKeyMu.Unlock()
	if !forceRefresh && wbiKeyCache != "" && time.Now().Before(wbiKeyExpireAt) {
		return wbiKeyCache, nil
	}
	key, err := fetchMixinKey(ctx)
	if err != nil {
		return "", err
	}
	wbiKeyCache, wbiKeyExpireAt = key, time.Now().Add(wbiKeyTTL)
	return key, nil
}

// fetchMixinKey 调 nav 接口推导签名密钥。
func fetchMixinKey(ctx context.Context) (string, error) {
	body, err := doGet(ctx, apiBaseURL+"/x/web-interface/nav", "")
	if err != nil {
		return "", fmt.Errorf("请求 nav 接口失败: %w", err)
	}
	var nav navResp
	if err := json.Unmarshal(body, &nav); err != nil {
		return "", fmt.Errorf("解析 nav 响应失败: %w", err)
	}
	raw := wbiURLKey(nav.Data.WbiImg.ImgURL) + wbiURLKey(nav.Data.WbiImg.SubURL)
	if len(raw) < mixinKeyLen {
		return "", fmt.Errorf("nav 响应缺少 wbi 密钥")
	}
	return deriveMixinKey(raw), nil
}

// getBuvidCookie 返回缓存的访客 Cookie（buvid3/buvid4），降低被风控概率；
// 获取失败返回空串（buvid 仅是缓释手段，不阻断解析）。
func getBuvidCookie(ctx context.Context) string {
	buvidMu.Lock()
	defer buvidMu.Unlock()
	if buvidCookieCache != "" && time.Now().Before(buvidExpireAt) {
		return buvidCookieCache
	}
	cookie, err := fetchBuvidCookie(ctx)
	if err != nil {
		return ""
	}
	buvidCookieCache, buvidExpireAt = cookie, time.Now().Add(buvidCookieTTL)
	return cookie
}

// fetchBuvidCookie 调 spi 接口获取 buvid3/buvid4 并拼成 Cookie 串。
func fetchBuvidCookie(ctx context.Context) (string, error) {
	body, err := doGet(ctx, apiBaseURL+"/x/frontend/finger/spi", "")
	if err != nil {
		return "", fmt.Errorf("请求 spi 接口失败: %w", err)
	}
	var spi spiResp
	if err := json.Unmarshal(body, &spi); err != nil {
		return "", fmt.Errorf("解析 spi 响应失败: %w", err)
	}
	if spi.Code != 0 || spi.Data.B3 == "" {
		return "", fmt.Errorf("获取 buvid 失败, code: %d", spi.Code)
	}
	return "buvid3=" + spi.Data.B3 + "; buvid4=" + spi.Data.B4, nil
}

// wbiURLKey 取 URL 中的文件名（去目录与扩展名），即 wbi img/sub key 本体。
func wbiURLKey(u string) string {
	if i := strings.LastIndex(u, "/"); i >= 0 {
		u = u[i+1:]
	}
	if i := strings.LastIndex(u, "."); i >= 0 {
		u = u[:i]
	}
	return u
}

// deriveMixinKey 按官方置换表从拼接的 img/sub key 推导 32 位签名密钥。
func deriveMixinKey(raw string) string {
	var b strings.Builder
	b.Grow(mixinKeyLen)
	for _, idx := range mixinKeyEncTab {
		if idx < len(raw) {
			b.WriteByte(raw[idx])
		}
	}
	key := b.String()
	if len(key) > mixinKeyLen {
		key = key[:mixinKeyLen]
	}
	return key
}

// signWBI 附加 wts 与 w_rid，返回按字典序排列的完整 query 串。
// 参数值须为无空格的 ASCII（当前仅 bvid 与数字），避免与
// JS encodeURIComponent 的转义差异影响签名结果。
func signWBI(params map[string]string, mixinKey string) string {
	return signWBIAt(params, time.Now().Unix(), mixinKey)
}

// signWBIAt signWBI 的可注入时间版本，便于测试与复现。
func signWBIAt(params map[string]string, wts int64, mixinKey string) string {
	signed := make(map[string]string, len(params)+1)
	for k, v := range params {
		signed[k] = filterWBIChars(v)
	}
	signed["wts"] = strconv.FormatInt(wts, 10)

	keys := make([]string, 0, len(signed))
	for k := range signed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(signed[k]))
	}
	query := strings.Join(parts, "&")
	sum := md5.Sum([]byte(query + mixinKey))
	return query + "&w_rid=" + hex.EncodeToString(sum[:])
}

// filterWBIChars 移除签名算法要求剔除的字符（!'()*）。
func filterWBIChars(s string) string {
	if !strings.ContainsAny(s, "!'()*") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if strings.ContainsRune("!'()*", r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
