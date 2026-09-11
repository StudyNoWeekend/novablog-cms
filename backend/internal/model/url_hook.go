package model

import "strings"

// resolveURL 将相对路径拼接为完整 URL。
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
