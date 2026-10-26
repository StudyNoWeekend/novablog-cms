package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/model"
)

// BloggerLogic 博主业务逻辑结构体。
type BloggerLogic struct {
	bloggerModel      *model.BloggerModel
	moduleConfigLogic *ModuleConfigLogic
}

// NewBloggerLogic 创建 BloggerLogic 实例。
func NewBloggerLogic() *BloggerLogic {
	return &BloggerLogic{
		bloggerModel:      model.NewBlogger(),
		moduleConfigLogic: NewModuleConfigLogic(),
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

// parseTags 将 JSON 格式的标签解析为字符串数组。
func parseTags(data string) []string {
	if data == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(data), &tags); err != nil {
		return []string{}
	}
	if tags == nil {
		return []string{}
	}
	return tags
}

// parseSocialLinksPublic 将 JSON 格式的社交链接解析并填充平台图标信息。
func parseSocialLinksPublic(data string) []res.SocialLinkPublicRes {
	links := parseSocialLinks(data)
	result := make([]res.SocialLinkPublicRes, 0, len(links))
	for _, link := range links {
		item := res.SocialLinkPublicRes{
			Platform:  link.Platform,
			URL:       link.URL,
			SortOrder: link.SortOrder,
		}
		if cfg, ok := GetSocialPlatformConfig(link.Platform); ok {
			item.Name = cfg.Name
			item.Icon = cfg.Icon
			item.Color = cfg.Color
		}
		result = append(result, item)
	}
	return result
}

// validateProfileMetaKey 校验性格/星座 key 白名单（空串合法，表示清除选择）。
func validateProfileMetaKey(kind, key string) error {
	if key == "" {
		return nil
	}
	switch kind {
	case metaKindPersonality:
		if _, ok := GetPersonalityConfig(key); ok {
			return nil
		}
		return enum.NewBizError(enum.ErrInvalidParam.Code, "未知的性格类型: "+key, enum.ErrInvalidParam.HttpCode)
	case metaKindZodiac:
		if _, ok := GetZodiacConfig(key); ok {
			return nil
		}
		return enum.NewBizError(enum.ErrInvalidParam.Code, "未知的星座: "+key, enum.ErrInvalidParam.HttpCode)
	}
	return nil
}

// buildZodiacPublicMeta 组装公开响应的星座信息，未设置或未开启对外展示时返回 nil。
func buildZodiacPublicMeta(show bool, key string) *res.ZodiacMetaRes {
	if !show || key == "" {
		return nil
	}
	cfg, ok := GetZodiacConfig(key)
	if !ok {
		return nil
	}
	return &res.ZodiacMetaRes{
		Key:       cfg.Key,
		Name:      cfg.Name,
		Image:     cfg.Image,
		DateRange: cfg.DateRange,
		Element:   cfg.Element,
	}
}

// buildPersonalityPublicMeta 组装公开响应的性格信息，未设置或未开启对外展示时返回 nil。
func buildPersonalityPublicMeta(show bool, key string) *res.PersonalityMetaRes {
	if !show || key == "" {
		return nil
	}
	cfg, ok := GetPersonalityConfig(key)
	if !ok {
		return nil
	}
	return &res.PersonalityMetaRes{
		Key:         cfg.Key,
		Name:        cfg.Name,
		Image:       cfg.Image,
		Description: cfg.Description,
	}
}

// GetPublicInfo 获取博主公开信息（排除敏感字段，并按对外展示开关脱敏）。
func (l *BloggerLogic) GetPublicInfo(ctx context.Context) (*res.BloggerPublicRes, error) {
	blogger, err := l.bloggerModel.GetFirst(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}
	l.resolveMediaURLs(ctx, blogger)

	// 邮箱 / 城市：未开启对外展示时返回空串
	email := ""
	if blogger.ShowEmail {
		email = blogger.Email
	}
	city := ""
	if blogger.ShowCity {
		city = blogger.City
	}

	return &res.BloggerPublicRes{
		Nickname:        blogger.Nickname,
		Avatar:          blogger.Avatar,
		Bio:             blogger.Bio,
		Email:           email,
		City:            city,
		Zodiac:          buildZodiacPublicMeta(blogger.ShowZodiac, blogger.Zodiac),
		Personality:     buildPersonalityPublicMeta(blogger.ShowPersonality, blogger.Personality),
		BlogTitle:       blogger.BlogTitle,
		BlogDescription: blogger.BlogDescription,
		PageBackground:  blogger.PageBackground,
		BlogIcon:        blogger.BlogIcon,
		SocialLinks:     parseSocialLinksPublic(blogger.SocialLinks),
		Tags:            parseTags(blogger.Tags),
	}, nil
}

// GetProfileMeta 获取个人资料选项元数据（星座 + 性格全量选项）。
func (l *BloggerLogic) GetProfileMeta() *res.ProfileMetaRes {
	zodiacList := ListZodiacConfigs()
	zodiacRes := make([]res.ZodiacMetaRes, 0, len(zodiacList))
	for _, cfg := range zodiacList {
		zodiacRes = append(zodiacRes, res.ZodiacMetaRes{
			Key:       cfg.Key,
			Name:      cfg.Name,
			Image:     cfg.Image,
			DateRange: cfg.DateRange,
			Element:   cfg.Element,
		})
	}

	personalityList := ListPersonalityConfigs()
	personalityRes := make([]res.PersonalityMetaRes, 0, len(personalityList))
	for _, cfg := range personalityList {
		personalityRes = append(personalityRes, res.PersonalityMetaRes{
			Key:         cfg.Key,
			Name:        cfg.Name,
			Image:       cfg.Image,
			Description: cfg.Description,
		})
	}

	return &res.ProfileMetaRes{
		Zodiac:      zodiacRes,
		Personality: personalityRes,
	}
}

// newProfileRes 博主模型 → 管理端资料响应 DTO。
func newProfileRes(blogger *model.Blogger) *res.BloggerProfileRes {
	return &res.BloggerProfileRes{
		Nickname:              blogger.Nickname,
		Avatar:                blogger.Avatar,
		AvatarMediaID:         refStr(blogger.AvatarMediaID),
		Bio:                   blogger.Bio,
		PageBackground:        blogger.PageBackground,
		PageBackgroundMediaID: refStr(blogger.PageBackgroundMediaID),
		BlogIcon:              blogger.BlogIcon,
		BlogIconMediaID:       refStr(blogger.BlogIconMediaID),
		BlogTitle:             blogger.BlogTitle,
		BlogDescription:       blogger.BlogDescription,
		Email:                 blogger.Email,
		City:                  blogger.City,
		Personality:           blogger.Personality,
		Zodiac:                blogger.Zodiac,
		ShowEmail:             blogger.ShowEmail,
		ShowCity:              blogger.ShowCity,
		ShowZodiac:            blogger.ShowZodiac,
		ShowPersonality:       blogger.ShowPersonality,
		Role:                  blogger.Role,
		SocialLinks:           parseSocialLinks(blogger.SocialLinks),
		Tags:                  parseTags(blogger.Tags),
	}
}

// resolveMediaURLs 解析博主头像/博客icon/页面背景的媒体引用为当前完整 URL（外链行保持原值）。
// 三列均无预设引用，按 media_id → 历史落库 URL 优先级解析。
func (l *BloggerLogic) resolveMediaURLs(ctx context.Context, blogger *model.Blogger) {
	blogger.Avatar = model.ResolveMediaRefURL(ctx, refStr(blogger.AvatarMediaID), "", blogger.Avatar)
	blogger.BlogIcon = model.ResolveMediaRefURL(ctx, refStr(blogger.BlogIconMediaID), "", blogger.BlogIcon)
	blogger.PageBackground = model.ResolveMediaRefURL(ctx, refStr(blogger.PageBackgroundMediaID), "", blogger.PageBackground)
}

// GetProfile 获取博主管理端资料信息。
func (l *BloggerLogic) GetProfile(ctx context.Context, userID string) (*res.BloggerProfileRes, error) {
	blogger, err := l.bloggerModel.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}
	l.resolveMediaURLs(ctx, blogger)

	return newProfileRes(blogger), nil
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
	if r.Avatar != nil || r.AvatarMediaID != nil {
		mediaID, presetID := "", ""
		if r.AvatarMediaID != nil {
			mediaID = *r.AvatarMediaID
		}
		avatarURL := ""
		if r.Avatar != nil {
			avatarURL = *r.Avatar
		}
		avatarMediaID, _, avatar, err := normalizeMediaRef(ctx, mediaID, presetID, avatarURL)
		if err != nil {
			return nil, err
		}
		blogger.Avatar = avatar
		blogger.AvatarMediaID = mediaRefPtr(avatarMediaID)
	}
	if r.Bio != nil {
		blogger.Bio = *r.Bio
	}
	if r.Email != nil {
		blogger.Email = *r.Email
	}
	if r.City != nil {
		blogger.City = *r.City
	}
	if r.Personality != nil {
		// 性格 key 白名单校验（空串表示清除）
		if err := validateProfileMetaKey(metaKindPersonality, *r.Personality); err != nil {
			return nil, err
		}
		blogger.Personality = *r.Personality
	}
	if r.Zodiac != nil {
		// 星座 key 白名单校验（空串表示清除）
		if err := validateProfileMetaKey(metaKindZodiac, *r.Zodiac); err != nil {
			return nil, err
		}
		blogger.Zodiac = *r.Zodiac
	}
	if r.ShowEmail != nil {
		blogger.ShowEmail = *r.ShowEmail
	}
	if r.ShowCity != nil {
		blogger.ShowCity = *r.ShowCity
	}
	if r.ShowZodiac != nil {
		blogger.ShowZodiac = *r.ShowZodiac
	}
	if r.ShowPersonality != nil {
		blogger.ShowPersonality = *r.ShowPersonality
	}
	if r.PageBackground != nil || r.PageBackgroundMediaID != nil {
		mediaID := ""
		if r.PageBackgroundMediaID != nil {
			mediaID = *r.PageBackgroundMediaID
		}
		bgURL := ""
		if r.PageBackground != nil {
			bgURL = *r.PageBackground
		}
		bgMediaID, _, bg, err := normalizeMediaRef(ctx, mediaID, "", bgURL)
		if err != nil {
			return nil, err
		}
		blogger.PageBackground = bg
		blogger.PageBackgroundMediaID = mediaRefPtr(bgMediaID)
	}
	if r.BlogIcon != nil || r.BlogIconMediaID != nil {
		mediaID := ""
		if r.BlogIconMediaID != nil {
			mediaID = *r.BlogIconMediaID
		}
		iconURL := ""
		if r.BlogIcon != nil {
			iconURL = *r.BlogIcon
		}
		iconMediaID, _, icon, err := normalizeMediaRef(ctx, mediaID, "", iconURL)
		if err != nil {
			return nil, err
		}
		blogger.BlogIcon = icon
		blogger.BlogIconMediaID = mediaRefPtr(iconMediaID)
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
	if r.Tags != nil {
		jsonBytes, err := json.Marshal(*r.Tags)
		if err != nil {
			return nil, fmt.Errorf("序列化标签失败: %w", err)
		}
		blogger.Tags = string(jsonBytes)
	}

	if err := l.bloggerModel.Update(ctx, blogger); err != nil {
		return nil, fmt.Errorf("更新博主信息失败: %w", err)
	}

	// ID 引用模式下落库 URL 为空，返回前解析为当前完整链接供前端回显
	l.resolveMediaURLs(ctx, blogger)
	return newProfileRes(blogger), nil
}

// UpdateRoles 补选创作方向角色（旧版本用户未选爱好的兼容入口，可重复提交）：
// 更新 bloggers.role，并按所选角色应用模块开关——显式 modules 全量覆盖优先，
// 未传时按角色预设并集兜底（通用模块 article/media 恒开，不受影响）。
func (l *BloggerLogic) UpdateRoles(ctx context.Context, userID string, r *req.UpdateRolesReq) (*res.BloggerProfileRes, error) {
	blogger, err := l.bloggerModel.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取博主信息失败: %w", err)
	}

	// 校验角色 key 白名单并去重
	seen := make(map[string]struct{}, len(r.Roles))
	roles := make([]string, 0, len(r.Roles))
	for _, key := range r.Roles {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := rolePresetModules[key]; !ok {
			return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "未知的创作方向角色: "+key, enum.ErrInvalidParam.HttpCode)
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		roles = append(roles, key)
	}
	if len(roles) == 0 {
		return nil, enum.ErrInvalidParam
	}

	// 校验模块开关 key 白名单
	for key := range r.Modules {
		if !IsValidModuleKey(key) {
			return nil, enum.NewBizError(enum.ErrInvalidParam.Code, "未知的模块开关: "+key, enum.ErrInvalidParam.HttpCode)
		}
	}

	// 应用模块开关预设（写入后 ApplyModuleMap 内部热更新缓存）
	if r.Modules != nil {
		if err := l.moduleConfigLogic.ApplyModuleMap(ctx, r.Modules); err != nil {
			return nil, err
		}
	} else if modules, presetErr := rolePreset(strings.Join(roles, ",")); presetErr == nil {
		if err := l.moduleConfigLogic.ApplyModuleMap(ctx, modules); err != nil {
			return nil, err
		}
	}

	blogger.Role = strings.Join(roles, ",")
	if err := l.bloggerModel.Update(ctx, blogger); err != nil {
		return nil, fmt.Errorf("更新创作方向失败: %w", err)
	}

	return newProfileRes(blogger), nil
}
