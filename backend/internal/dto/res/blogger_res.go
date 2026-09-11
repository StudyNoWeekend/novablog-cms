package res

// SocialLinkRes 社交平台链接响应结构体。
type SocialLinkRes struct {
	Platform  string `json:"platform"`   // 平台标识
	URL       string `json:"url"`        // 个人主页 URL
	SortOrder int    `json:"sort_order"` // 排序权重
}

// SocialLinkPublicRes 公开接口社交平台链接响应结构体（包含图标信息）。
type SocialLinkPublicRes struct {
	Platform  string `json:"platform"`   // 平台标识
	Name      string `json:"name"`       // 平台名称
	Icon      string `json:"icon"`       // SVG 图标 path 数据
	Color     string `json:"color"`      // 品牌色
	URL       string `json:"url"`        // 个人主页 URL
	SortOrder int    `json:"sort_order"` // 排序权重
}

// BloggerPublicRes 博主公开信息响应结构体。
type BloggerPublicRes struct {
	Nickname        string                `json:"nickname"`         // 昵称
	Avatar          string                `json:"avatar"`           // 头像 URL
	Bio             string                `json:"bio"`              // 个人简介
	BlogTitle       string                `json:"blog_title"`       // 博客标题
	BlogDescription string                `json:"blog_description"` // 博客描述
	PageBackground  string                `json:"page_background"`  // 页面背景图 URL
	BlogIcon        string                `json:"blog_icon"`        // 博客 icon 图 URL
	SocialLinks     []SocialLinkPublicRes `json:"social_links"`     // 社交平台链接数组（含图标）
	Tags            []string              `json:"tags"`             // 标签数组
}

// BloggerProfileRes 博主管理端资料响应结构体。
type BloggerProfileRes struct {
	Nickname        string          `json:"nickname"`         // 昵称
	Avatar          string          `json:"avatar"`           // 头像 URL
	Bio             string          `json:"bio"`              // 个人简介
	PageBackground  string          `json:"page_background"`  // 页面背景图 URL
	BlogIcon        string          `json:"blog_icon"`        // 博客 icon 图 URL
	BlogTitle       string          `json:"blog_title"`       // 博客标题
	BlogDescription string          `json:"blog_description"` // 博客描述
	Email           string          `json:"email"`            // 邮箱
	SocialLinks     []SocialLinkRes `json:"social_links"`     // 社交平台链接数组
	Tags            []string        `json:"tags"`             // 标签数组
}
