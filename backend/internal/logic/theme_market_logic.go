package logic

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/pkg/novablogapi"
)

// ThemeMarketLogic 官方主题市场转发逻辑。
//
// 官方地址与 Token 由控制器从请求头读出后传入（请求头为空时回退后端配置）；
// 登录成功后官方地址持久化为全局默认，其余官方信息不落库、不缓存。
type ThemeMarketLogic struct{}

// NewThemeMarketLogic 创建 ThemeMarketLogic 实例。
func NewThemeMarketLogic() *ThemeMarketLogic {
	return &ThemeMarketLogic{}
}

// normalizeBaseURL 归一化官方地址，失败时转为业务错误。
// 请求头未携带地址时回退后端配置（DB 持久化值优先，出厂值为 config.yaml）。
func normalizeBaseURL(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		raw = getThemeSettings().MarketBaseURL
	}
	base, err := novablogapi.NormalizeBaseURL(raw)
	if err != nil {
		return "", enum.NewBizError(enum.ErrMarketBaseURLInvalid.Code, err.Error(), enum.ErrMarketBaseURLInvalid.HttpCode)
	}
	return base, nil
}

// mapUpstreamError 将官方客户端错误映射为业务错误。
// 官方 401 → 登录失效；官方业务错误 → 保留官方描述透传；其余 → 上游不可用。
func mapUpstreamError(err error) error {
	var authErr *novablogapi.AuthError
	var apiErr *novablogapi.APIError
	switch {
	case errors.As(err, &authErr):
		return enum.NewBizError(enum.ErrMarketAuthFailed.Code, enum.ErrMarketAuthFailed.Msg, enum.ErrMarketAuthFailed.HttpCode)
	case errors.As(err, &apiErr):
		return enum.NewBizError(apiErr.Code, apiErr.Message, enum.ErrMarketBaseURLInvalid.HttpCode)
	default:
		return enum.NewBizError(enum.ErrMarketUpstream.Code, enum.ErrMarketUpstream.Msg, enum.ErrMarketUpstream.HttpCode)
	}
}

// mapThemeItem 官方主题条目 → 响应 DTO。
func mapThemeItem(item novablogapi.ThemeItem) res.ThemeMarketItemRes {
	styles := item.Styles
	if styles == nil {
		styles = []string{}
	}
	features := item.Features
	if features == nil {
		features = []string{}
	}
	return res.ThemeMarketItemRes{
		ID:          item.ID,
		Title:       item.Title,
		Slug:        item.Slug,
		Description: item.Description,
		AuthorID:    item.AuthorID,
		Author:      item.Author,
		Price:       item.Price,
		PriceAmount: item.PriceAmount,
		Type:        item.Type,
		Styles:      styles,
		Features:    features,
		Preview:     item.Preview,
		Version:     item.Version,
		Downloads:   item.Downloads,
		Likes:       item.Likes,
		Rating:      item.Rating,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt,
	}
}

// mapThemeList 官方分页列表 → 统一分页响应。
func mapThemeList(list novablogapi.ThemeList, page, pageSize int) *res.PageRes[res.ThemeMarketItemRes] {
	items := make([]res.ThemeMarketItemRes, 0, len(list.Items))
	for _, it := range list.Items {
		items = append(items, mapThemeItem(it))
	}
	return res.NewPageRes(items, list.Total, page, pageSize)
}

// MarketLogin 登录官方主题市场，返回官方双 Token 与用户信息。
func (l *ThemeMarketLogic) MarketLogin(ctx context.Context, baseURL string, r *req.MarketLoginReq) (*res.MarketLoginRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	pair, user, err := novablogapi.Login(ctx, base, r.Email, r.Password)
	if err != nil {
		// 登录失败透传官方文案（如"邮箱或密码错误"），便于前端就地提示
		var authErr *novablogapi.AuthError
		if errors.As(err, &authErr) && authErr.Message != "" {
			return nil, enum.NewBizError(enum.ErrMarketLoginFailed.Code, authErr.Message, enum.ErrMarketLoginFailed.HttpCode)
		}
		return nil, mapUpstreamError(err)
	}

	// 登录成功后将官方地址持久化为全局默认（后台模板市场页可修改），失败不影响登录结果
	if err := NewThemeMarketConfigLogic().SyncMarketBaseURL(ctx, base); err != nil {
		themeLog().Warn("官方地址持久化失败", zap.Error(err))
	}

	return &res.MarketLoginRes{
		AccessToken: pair.AccessToken,
		ExpiresIn:   pair.ExpiresIn,
		User: res.MarketUserRes{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Role:     user.Role,
		},
	}, nil
}

// MarketLogout 登出官方账号（Token 拉黑）。Token 已失效时视为成功，便于本地退出。
func (l *ThemeMarketLogic) MarketLogout(ctx context.Context, baseURL, token string) error {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return err
	}

	if err := novablogapi.Logout(ctx, base, token); err != nil {
		var authErr *novablogapi.AuthError
		if errors.As(err, &authErr) {
			return nil
		}
		return mapUpstreamError(err)
	}
	return nil
}

// GetList 查询官方主题市场列表（仅已上架）。
func (l *ThemeMarketLogic) GetList(ctx context.Context, baseURL, token string, r *req.ThemeMarketListReq) (*res.PageRes[res.ThemeMarketItemRes], error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	list, err := novablogapi.ListThemes(ctx, base, token, novablogapi.ThemeQuery{
		Page:     r.GetPage(),
		Size:     r.GetPageSize(),
		Type:     r.Type,
		Styles:   r.Styles,
		Features: r.Features,
		Price:    r.Price,
		Search:   r.Search,
		Sort:     r.Sort,
	})
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return mapThemeList(list, r.GetPage(), r.GetPageSize()), nil
}

// GetDetail 查询主题详情（携带官方 Token 时返回个人点赞/评分状态）。
func (l *ThemeMarketLogic) GetDetail(ctx context.Context, baseURL, token, id string) (*res.ThemeMarketDetailRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	detail, err := novablogapi.GetTheme(ctx, base, id, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	return &res.ThemeMarketDetailRes{
		ThemeMarketItemRes: mapThemeItem(detail.ThemeItem),
		DownloadURL:        detail.DownloadURL,
		Liked:              detail.Liked,
		UserRating:         detail.UserRating,
	}, nil
}

// GetReleases 查询主题版本历史与更新日志（按发布时间倒序，无记录返回空数组）。
func (l *ThemeMarketLogic) GetReleases(ctx context.Context, baseURL, token, id string) ([]res.ThemeMarketReleaseRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	items, err := novablogapi.GetThemeReleases(ctx, base, id, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	result := make([]res.ThemeMarketReleaseRes, 0, len(items))
	for _, it := range items {
		result = append(result, res.ThemeMarketReleaseRes{
			Version:     it.Version,
			Tag:         it.Tag,
			Notes:       it.Notes,
			PublishedAt: it.PublishedAt,
		})
	}
	return result, nil
}

// GetStats 查询市场统计。
func (l *ThemeMarketLogic) GetStats(ctx context.Context, baseURL string) (*res.ThemeMarketStatsRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	stats, err := novablogapi.GetStats(ctx, base)
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	return &res.ThemeMarketStatsRes{
		Total:     stats.Total,
		Authors:   stats.Authors,
		Downloads: stats.Downloads,
	}, nil
}

// GetHotTags 查询热门风格标签。
func (l *ThemeMarketLogic) GetHotTags(ctx context.Context, baseURL string) ([]res.ThemeMarketHotTagRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	tags, err := novablogapi.GetHotTags(ctx, base)
	if err != nil {
		return nil, mapUpstreamError(err)
	}

	result := make([]res.ThemeMarketHotTagRes, 0, len(tags))
	for _, t := range tags {
		result = append(result, res.ThemeMarketHotTagRes{Name: t.Name, Count: t.Count})
	}
	return result, nil
}

// Like 点赞/取消点赞官方主题（toggle）。
func (l *ThemeMarketLogic) Like(ctx context.Context, baseURL, token, id string) (*res.ThemeMarketLikeRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	result, err := novablogapi.LikeTheme(ctx, base, id, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return &res.ThemeMarketLikeRes{Liked: result.Liked}, nil
}

// Favorite 收藏/取消收藏官方主题（toggle）。
func (l *ThemeMarketLogic) Favorite(ctx context.Context, baseURL, token, id string) (*res.ThemeMarketFavoriteRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	result, err := novablogapi.FavoriteTheme(ctx, base, id, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return &res.ThemeMarketFavoriteRes{Favorited: result.Favorited}, nil
}

// Rating 为官方主题评分（1-5 分，覆盖式）。
func (l *ThemeMarketLogic) Rating(ctx context.Context, baseURL, token, id string, r *req.ThemeMarketRatingReq) (*res.ThemeMarketRatingRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	result, err := novablogapi.RateTheme(ctx, base, id, r.Score, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return &res.ThemeMarketRatingRes{Rating: result.Rating}, nil
}

// Download 下载官方主题：请求官方代理下载接口（计数 +1）并解析出真实下载地址，
// 前端拿到 GitHub 直链后可直接新窗口打开下载，无需携带官方 Token。
func (l *ThemeMarketLogic) Download(ctx context.Context, baseURL, token, id string) (*res.ThemeMarketDownloadRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(token) == "" {
		return nil, enum.ErrMarketAuthFailed
	}
	location, err := novablogapi.ResolveDownload(ctx, base+"/themes/"+id+"/download", token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return &res.ThemeMarketDownloadRes{DownloadURL: location}, nil
}

// GetFavorites 查询我的收藏列表（按收藏时间倒序）。
func (l *ThemeMarketLogic) GetFavorites(ctx context.Context, baseURL, token string, r *req.ThemeMarketPageReq) (*res.PageRes[res.ThemeMarketItemRes], error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	list, err := novablogapi.ListFavorites(ctx, base, token, r.GetPage(), r.GetPageSize())
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return mapThemeList(list, r.GetPage(), r.GetPageSize()), nil
}

// GetFavoriteIds 查询收藏主题 ID 集合。
func (l *ThemeMarketLogic) GetFavoriteIds(ctx context.Context, baseURL, token string) ([]int64, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	ids, err := novablogapi.ListFavoriteIDs(ctx, base, token)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	if ids == nil {
		ids = []int64{}
	}
	return ids, nil
}

// GetDefault 查询官方默认主题（部署首装直接拉取；未设置或已下架时官方返回 404）。
func (l *ThemeMarketLogic) GetDefault(ctx context.Context, baseURL string) (*res.ThemeMarketDefaultRes, error) {
	base, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	dt, err := novablogapi.GetDefaultTheme(ctx, base)
	if err != nil {
		return nil, mapUpstreamError(err)
	}
	return &res.ThemeMarketDefaultRes{
		ThemeMarketItemRes: mapThemeItem(dt.ThemeItem),
		IsDefault:          dt.IsDefault,
		DownloadURL:        dt.DownloadURL,
	}, nil
}
