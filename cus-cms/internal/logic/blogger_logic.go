package logic

import (
	"context"
	"fmt"

	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"
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
	}, nil
}
