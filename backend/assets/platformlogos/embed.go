// Package platformlogos 内嵌第三方音乐平台品牌 Logo（SVG，白色填充，24x24 viewBox）。
// 歌单未设置封面时由后端按 platform 下发对应 Logo，前端与博客主题无需内置品牌资源。
// 素材来源：Simple Icons（CC0）与自绘音符；品牌图形版权归各平台所有。
package platformlogos

import (
	"embed"
)

//go:embed *.svg
var files embed.FS

// logoFiles 平台标识 -> 内嵌 SVG 文件名，与数据库 platform 字段枚举一致。
var logoFiles = map[string]string{
	"qq_music":    "qq_music.svg",
	"netease":     "netease.svg",
	"bilibili":    "bilibili.svg",
	"spotify":     "spotify.svg",
	"apple_music": "apple_music.svg",
	"other":       "other.svg",
}

// Has 返回平台是否内置 Logo。
func Has(platform string) bool {
	_, ok := logoFiles[platform]
	return ok
}

// Get 返回平台 Logo 的 SVG 内容；平台未知时返回 false。
func Get(platform string) ([]byte, bool) {
	name, ok := logoFiles[platform]
	if !ok {
		return nil, false
	}
	data, err := files.ReadFile(name)
	if err != nil {
		return nil, false
	}
	return data, true
}

// URLPath 返回平台 Logo 的公开访问路径（不含 BaseURL 前缀），与 public.go 的路由保持一致。
// 平台未知时返回空串。
func URLPath(platform string) string {
	name, ok := logoFiles[platform]
	if !ok {
		return ""
	}
	return "/api/v1/public/platform-logos/" + name
}
