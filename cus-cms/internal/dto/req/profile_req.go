package req

// UpdateProfileReq 更新个人资料请求。
type UpdateProfileReq struct {
	Nickname        *string `json:"nickname" binding:"omitempty,max=50"`         // 昵称
	Avatar          *string `json:"avatar" binding:"omitempty,max=500"`          // 头像 URL
	Bio             *string `json:"bio" binding:"omitempty"`                     // 个人简介
	PageBackground  *string `json:"page_background" binding:"omitempty,max=500"` // 页面背景图 URL
	BlogIcon        *string `json:"blog_icon" binding:"omitempty,max=500"`       // 博客 icon 图 URL
	BlogTitle       *string `json:"blog_title" binding:"omitempty,max=100"`      // 博客标题
	BlogDescription *string `json:"blog_description" binding:"omitempty"`        // 博客描述
}
