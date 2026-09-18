package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// 官方主题市场转发专用请求头。
const (
	// HeaderMarketBaseURL 官方服务地址（由前端登录弹窗输入并随请求携带）。
	HeaderMarketBaseURL = "X-Market-Base-URL"
	// HeaderMarketToken 官方账号 Access Token（由前端登录后随请求携带）。
	HeaderMarketToken = "X-Market-Token"
)

// ThemeMarketController 官方主题市场控制器（无状态转发代理）。
type ThemeMarketController struct {
	logic *logic.ThemeMarketLogic
}

// NewThemeMarketController 创建 ThemeMarketController 实例。
func NewThemeMarketController() *ThemeMarketController {
	return &ThemeMarketController{logic: logic.NewThemeMarketLogic()}
}

// marketBaseURL 读取官方服务地址请求头。
func (c *ThemeMarketController) marketBaseURL(ctx *gin.Context) string {
	return ctx.GetHeader(HeaderMarketBaseURL)
}

// marketToken 读取官方 Token 请求头。
func (c *ThemeMarketController) marketToken(ctx *gin.Context) string {
	return ctx.GetHeader(HeaderMarketToken)
}

// MarketLogin 登录官方主题市场 POST /api/v1/themes/market/auth/login
func (c *ThemeMarketController) MarketLogin(ctx *gin.Context) {
	var r req.MarketLoginReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, "请输入正确的官方地址、邮箱和密码", enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.MarketLogin(ctx.Request.Context(), c.marketBaseURL(ctx), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// MarketLogout 登出官方账号 POST /api/v1/themes/market/auth/logout
func (c *ThemeMarketController) MarketLogout(ctx *gin.Context) {
	if err := c.logic.MarketLogout(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx)); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// GetList 官方主题市场列表 GET /api/v1/themes/market
func (c *ThemeMarketController) GetList(ctx *gin.Context) {
	var r req.ThemeMarketListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetList(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetDetail 官方主题详情 GET /api/v1/themes/market/detail/:id
func (c *ThemeMarketController) GetDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetDetail(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetReleases 官方主题版本历史与更新日志 GET /api/v1/themes/market/:id/releases
func (c *ThemeMarketController) GetReleases(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetReleases(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetStats 官方市场统计 GET /api/v1/themes/market/stats
func (c *ThemeMarketController) GetStats(ctx *gin.Context) {
	result, err := c.logic.GetStats(ctx.Request.Context(), c.marketBaseURL(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetHotTags 官方热门风格标签 GET /api/v1/themes/market/hot-tags
func (c *ThemeMarketController) GetHotTags(ctx *gin.Context) {
	result, err := c.logic.GetHotTags(ctx.Request.Context(), c.marketBaseURL(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetDefault 官方默认主题 GET /api/v1/themes/market/default
func (c *ThemeMarketController) GetDefault(ctx *gin.Context) {
	result, err := c.logic.GetDefault(ctx.Request.Context(), c.marketBaseURL(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetFavorites 我的收藏列表 GET /api/v1/themes/market/favorites
func (c *ThemeMarketController) GetFavorites(ctx *gin.Context) {
	var r req.ThemeMarketPageReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetFavorites(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetFavoriteIds 收藏主题 ID 集合 GET /api/v1/themes/market/favorites/ids
func (c *ThemeMarketController) GetFavoriteIds(ctx *gin.Context) {
	result, err := c.logic.GetFavoriteIds(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Like 点赞/取消点赞 POST /api/v1/themes/market/:id/like
func (c *ThemeMarketController) Like(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.Like(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Favorite 收藏/取消收藏 POST /api/v1/themes/market/:id/favorite
func (c *ThemeMarketController) Favorite(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.Favorite(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Rating 评分 POST /api/v1/themes/market/:id/rating
func (c *ThemeMarketController) Rating(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.ThemeMarketRatingReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, "评分需为 1-5 的整数", enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Rating(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Download 下载/安装 POST /api/v1/themes/market/:id/download
func (c *ThemeMarketController) Download(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.Download(ctx.Request.Context(), c.marketBaseURL(ctx), c.marketToken(ctx), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}
