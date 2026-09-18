package req

// SocialLinkReq 社交平台链接请求结构体。
type SocialLinkReq struct {
	Platform  string `json:"platform" binding:"required,max=50"` // 平台标识
	URL       string `json:"url" binding:"required,max=500"`     // 个人主页 URL
	SortOrder int    `json:"sort_order"`                         // 排序权重
}

// UpdateProfileReq 更新个人资料请求。
type UpdateProfileReq struct {
	Nickname        *string          `json:"nickname" binding:"omitempty,max=50"`         // 昵称
	Avatar          *string          `json:"avatar" binding:"omitempty,max=500"`          // 头像 URL
	Bio             *string          `json:"bio" binding:"omitempty"`                     // 个人简介
	Email           *string          `json:"email" binding:"omitempty,max=100"`           // 邮箱
	City            *string          `json:"city" binding:"omitempty,max=100"`            // 所在城市
	PageBackground  *string          `json:"page_background" binding:"omitempty,max=500"` // 页面背景图 URL
	BlogIcon        *string          `json:"blog_icon" binding:"omitempty,max=500"`       // 博客 icon 图 URL
	BlogTitle       *string          `json:"blog_title" binding:"omitempty,max=100"`      // 博客标题
	BlogDescription *string          `json:"blog_description" binding:"omitempty"`        // 博客描述
	SocialLinks     *[]SocialLinkReq `json:"social_links" binding:"omitempty"`            // 社交平台链接数组
	Tags            *[]string        `json:"tags" binding:"omitempty,max=20,dive,max=30"` // 标签数组
}
