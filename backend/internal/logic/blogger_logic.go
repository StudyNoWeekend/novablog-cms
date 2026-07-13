package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// BloggerLogic 博主业务逻辑结构体。
type BloggerLogic struct {
	bloggerModel *model.BloggerModel
}

// NewBloggerLogic 创建 BloggerLogic 实例。
func NewBloggerLogic() *BloggerLogic {
	return &BloggerLogic{
		bloggerModel: model.NewBlogger(),
	}
}

// parseSocialLinks 将 JSON 格式的社交链接解析为响应结构体数组。
func parseSocialLinks(data string) []res.SocialLinkRes {
	if data == "" {
		return []res.SocialLinkRes{}
	}
	var links []res.SocialLinkRes
	if err := json.Unmarshal([]byte(data), &links); err != nil {
		return []res.SocialLinkRes{}
	}
	if links == nil {
		return []res.SocialLinkRes{}
	}
	return links
}

// GetPublicInfo 获取博主公开信息（排除敏感字段）。
func (l *BloggerLogic) GetPublicInfo(ctx context.Context) (*res.BloggerPublicRes, error) {
	blogger, err := l.bloggerModel.GetFirst(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}

	return &res.BloggerPublicRes{
		Nickname:        blogger.Nickname,
		Avatar:          blogger.Avatar,
		Bio:             blogger.Bio,
		BlogTitle:       blogger.BlogTitle,
		BlogDescription: blogger.BlogDescription,
		PageBackground:  blogger.PageBackground,
		BlogIcon:        blogger.BlogIcon,
		SocialLinks:     parseSocialLinks(blogger.SocialLinks),
	}, nil
}

// GetProfile 获取博主管理端资料信息。
func (l *BloggerLogic) GetProfile(ctx context.Context, userID string) (*res.BloggerProfileRes, error) {
	blogger, err := l.bloggerModel.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}

	return &res.BloggerProfileRes{
		Nickname:        blogger.Nickname,
		Avatar:          blogger.Avatar,
		Bio:             blogger.Bio,
		PageBackground:  blogger.PageBackground,
		BlogIcon:        blogger.BlogIcon,
		BlogTitle:       blogger.BlogTitle,
		BlogDescription: blogger.BlogDescription,
		Email:           blogger.Email,
		SocialLinks:     parseSocialLinks(blogger.SocialLinks),
	}, nil
}

// UpdateProfile 更新博主个人资料。
func (l *BloggerLogic) UpdateProfile(ctx context.Context, userID string, r *req.UpdateProfileReq) (*res.BloggerProfileRes, error) {
	blogger, err := l.bloggerModel.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}

	if r.Nickname != nil {
		blogger.Nickname = *r.Nickname
	}
	if r.Avatar != nil {
		blogger.Avatar = *r.Avatar
	}
	if r.Bio != nil {
		blogger.Bio = *r.Bio
	}
	if r.PageBackground != nil {
		blogger.PageBackground = *r.PageBackground
	}
	if r.BlogIcon != nil {
		blogger.BlogIcon = *r.BlogIcon
	}
	if r.BlogTitle != nil {
		blogger.BlogTitle = *r.BlogTitle
	}
	if r.BlogDescription != nil {
		blogger.BlogDescription = *r.BlogDescription
	}
	if r.SocialLinks != nil {
		jsonBytes, err := json.Marshal(*r.SocialLinks)
		if err != nil {
			return nil, fmt.Errorf("序列化社交链接失败: %w", err)
		}
		blogger.SocialLinks = string(jsonBytes)
	}

	if err := l.bloggerModel.Update(ctx, blogger); err != nil {
		return nil, fmt.Errorf("更新博主信息失败: %w", err)
	}

	return &res.BloggerProfileRes{
		Nickname:        blogger.Nickname,
		Avatar:          blogger.Avatar,
		Bio:             blogger.Bio,
		PageBackground:  blogger.PageBackground,
		BlogIcon:        blogger.BlogIcon,
		BlogTitle:       blogger.BlogTitle,
		BlogDescription: blogger.BlogDescription,
		Email:           blogger.Email,
		SocialLinks:     parseSocialLinks(blogger.SocialLinks),
	}, nil
}
