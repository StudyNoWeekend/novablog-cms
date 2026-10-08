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

// ZodiacMetaRes 星座展示信息响应结构体。
type ZodiacMetaRes struct {
	Key       string `json:"key"`        // 星座 key
	Name      string `json:"name"`       // 星座名称
	Image     string `json:"image"`      // 星座连线图 SVG 片段（24×24 viewBox 内）
	DateRange string `json:"date_range"` // 日期范围
	Element   string `json:"element"`    // 星座类型（火象/土象/风象/水象）
}

// PersonalityMetaRes 性格展示信息响应结构体（MBTI）。
type PersonalityMetaRes struct {
	Key         string `json:"key"`         // MBTI 类型 key
	Name        string `json:"name"`        // 类型名称
	Image       string `json:"image"`       // 四字母徽章 SVG 片段（24×24 viewBox 内）
	Description string `json:"description"` // 一句话介绍
}

// ProfileMetaRes 个人资料选项元数据响应结构体（星座 + 性格全量选项）。
type ProfileMetaRes struct {
	Zodiac      []ZodiacMetaRes      `json:"zodiac"`      // 星座选项数组
	Personality []PersonalityMetaRes `json:"personality"` // 性格选项数组
}

// BloggerPublicRes 博主公开信息响应结构体。
type BloggerPublicRes struct {
	Nickname        string                `json:"nickname"`         // 昵称
	Avatar          string                `json:"avatar"`           // 头像 URL
	Bio             string                `json:"bio"`              // 个人简介
	Email           string                `json:"email"`            // 邮箱（未开启对外展示时为空串）
	City            string                `json:"city"`             // 所在城市（未开启对外展示时为空串）
	Zodiac          *ZodiacMetaRes        `json:"zodiac"`           // 星座信息（未设置或未开启对外展示时为 null）
	Personality     *PersonalityMetaRes   `json:"personality"`      // 性格信息（未设置或未开启对外展示时为 null）
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
	City            string          `json:"city"`             // 所在城市
	Personality     string          `json:"personality"`      // 性格（MBTI key，空串表示未设置）
	Zodiac          string          `json:"zodiac"`           // 星座 key（空串表示未设置）
	ShowEmail       bool            `json:"show_email"`       // 邮箱是否对外展示
	ShowCity        bool            `json:"show_city"`        // 城市是否对外展示
	ShowZodiac      bool            `json:"show_zodiac"`      // 星座是否对外展示
	ShowPersonality bool            `json:"show_personality"` // 性格是否对外展示
	Role            string          `json:"role"`             // 创作方向（首装/补选所选角色 key，多选逗号分隔；空串表示未选过）
	SocialLinks     []SocialLinkRes `json:"social_links"`     // 社交平台链接数组
	Tags            []string        `json:"tags"`             // 标签数组
}
