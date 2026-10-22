package mcp

import (
	"context"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetProfileIn get_profile 入参（无参数）。
type GetProfileIn struct{}

// getProfile 读取博主资料（关于我页面的全部信息）。博主为站点唯一单行记录，
// MCP 无 JWT 身份，直接以唯一博主代替 JWT user_id。
func (s *Service) getProfile(ctx context.Context, _ *sdk.CallToolRequest, _ GetProfileIn) (*sdk.CallToolResult, res.BloggerProfileRes, error) {
	profile, err := s.bloggerLogic.GetProfile(ctx, s.currentBloggerID(ctx))
	if err != nil {
		return nil, res.BloggerProfileRes{}, fmt.Errorf("读取博主资料失败: %w", err)
	}
	return nil, *profile, nil
}

// ProfileSocialLinkIn update_profile 入参中的社交链接项。
type ProfileSocialLinkIn struct {
	Platform  string `json:"platform" jsonschema:"平台标识（如 github/weibo）"`
	URL       string `json:"url" jsonschema:"个人主页 URL"`
	SortOrder int    `json:"sort_order" jsonschema:"排序权重，小的在前"`
}

// UpdateProfileIn update_profile 入参（指针字段=不传不修改）。
type UpdateProfileIn struct {
	Nickname        *string                `json:"nickname,omitempty" jsonschema:"昵称"`
	Avatar          *string                `json:"avatar,omitempty" jsonschema:"头像 URL"`
	Bio             *string                `json:"bio,omitempty" jsonschema:"个人简介"`
	Email           *string                `json:"email,omitempty" jsonschema:"邮箱"`
	City            *string                `json:"city,omitempty" jsonschema:"所在城市"`
	Personality     *string                `json:"personality,omitempty" jsonschema:"性格 MBTI key（白名单校验），传空串清除"`
	Zodiac          *string                `json:"zodiac,omitempty" jsonschema:"星座 key（白名单校验），传空串清除"`
	ShowEmail       *bool                  `json:"show_email,omitempty" jsonschema:"邮箱是否对外展示"`
	ShowCity        *bool                  `json:"show_city,omitempty" jsonschema:"城市是否对外展示"`
	ShowZodiac      *bool                  `json:"show_zodiac,omitempty" jsonschema:"星座是否对外展示"`
	ShowPersonality *bool                  `json:"show_personality,omitempty" jsonschema:"性格是否对外展示"`
	PageBackground  *string                `json:"page_background,omitempty" jsonschema:"页面背景图 URL"`
	BlogIcon        *string                `json:"blog_icon,omitempty" jsonschema:"博客 icon 图 URL"`
	BlogTitle       *string                `json:"blog_title,omitempty" jsonschema:"博客标题"`
	BlogDescription *string                `json:"blog_description,omitempty" jsonschema:"博客描述"`
	SocialLinks     *[]ProfileSocialLinkIn `json:"social_links,omitempty" jsonschema:"社交平台链接数组（整体替换，传空数组表示清空）"`
	Tags            *[]string              `json:"tags,omitempty" jsonschema:"个人标签数组（整体替换，传空数组表示清空，最多 20 个）"`
}

// updateProfile 更新博主资料（PATCH 语义）。
func (s *Service) updateProfile(ctx context.Context, _ *sdk.CallToolRequest, in UpdateProfileIn) (*sdk.CallToolResult, res.BloggerProfileRes, error) {
	patch := &req.UpdateProfileReq{
		Nickname:        in.Nickname,
		Avatar:          in.Avatar,
		Bio:             in.Bio,
		Email:           in.Email,
		City:            in.City,
		Personality:     in.Personality,
		Zodiac:          in.Zodiac,
		ShowEmail:       in.ShowEmail,
		ShowCity:        in.ShowCity,
		ShowZodiac:      in.ShowZodiac,
		ShowPersonality: in.ShowPersonality,
		PageBackground:  in.PageBackground,
		BlogIcon:        in.BlogIcon,
		BlogTitle:       in.BlogTitle,
		BlogDescription: in.BlogDescription,
		Tags:            in.Tags,
	}
	if in.SocialLinks != nil {
		links := make([]req.SocialLinkReq, 0, len(*in.SocialLinks))
		for _, l := range *in.SocialLinks {
			if l.Platform == "" || l.URL == "" {
				return nil, res.BloggerProfileRes{}, fmt.Errorf("social_links 中 platform 与 url 均为必填")
			}
			links = append(links, req.SocialLinkReq{Platform: l.Platform, URL: l.URL, SortOrder: l.SortOrder})
		}
		patch.SocialLinks = &links
	}

	profile, err := s.bloggerLogic.UpdateProfile(ctx, s.currentBloggerID(ctx), patch)
	if err != nil {
		return nil, res.BloggerProfileRes{}, fmt.Errorf("更新博主资料失败: %w", err)
	}
	return nil, *profile, nil
}

// currentBloggerID 返回站点唯一博主的 ID。MCP 调用无 JWT 身份，
// CommentLogic.Reply / BloggerLogic.UpdateProfile 等以博主身份操作的 logic 需要它。
func (s *Service) currentBloggerID(ctx context.Context) string {
	b, err := model.NewBlogger().GetFirst(ctx)
	if err != nil {
		return ""
	}
	return b.ID
}
