package model

import "strings"

// ResolveStorageURL 由 bootstrap 启动时注入（实现为 storage.Manager.ResolvePublicURL）：
// 按存储类型与相对对象键动态拼接当前配置下可访问的完整 URL，无法解析时返回空串。
// model 包不能反向 import storage 包（storage 依赖 model），故以函数变量注入。
var ResolveStorageURL func(storageType, key string) string

// resolveURL 将本地存储的相对路径拼接为完整 URL。
// 如果 URL 为空或已以 http 开头，则原样返回。
// 否则拼接 BaseURL + "/files/" + path。
func resolveURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	if BaseURL == "" {
		return url
	}
	return strings.TrimRight(BaseURL, "/") + "/files/" + url
}

// ResolveURL 导出的 URL 解析函数，供 logic 层调用。
func ResolveURL(url string) string {
	return resolveURL(url)
}

// ResolveMediaURL 导出的媒体 URL 解析函数（resolveMediaURL），供 logic 层解析
// 未经 AfterFind 钩子的内存对象（如刚 Create 的记录）。
func ResolveMediaURL(storageType, raw string) string {
	return resolveMediaURL(storageType, raw)
}

// resolveMediaURL 统一解析媒体相关 URL（media.url / media_presets.output_url / thumb）：
//   - 空 或 已是完整 http(s) URL（历史落库值/外链）→ 原样返回；
//   - 本地存储（或未知类型）→ resolveURL 拼 BaseURL + /files/；
//   - 对象存储 → ResolveStorageURL 按当前平台配置动态拼接，失败回退原值。
func resolveMediaURL(storageType, raw string) string {
	if raw == "" || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if storageType == "" || storageType == "local" {
		return resolveURL(raw)
	}
	if ResolveStorageURL != nil {
		if resolved := ResolveStorageURL(storageType, raw); resolved != "" {
			return resolved
		}
	}
	return raw
}
