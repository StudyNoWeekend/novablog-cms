package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ThemeController 已安装主题管理控制器（安装/激活/卸载/列表）。
type ThemeController struct {
	logic *logic.ThemeLogic
}

// NewThemeController 创建 ThemeController 实例。
func NewThemeController() *ThemeController {
	return &ThemeController{logic: logic.NewThemeLogic()}
}

// marketBaseURL 读取官方服务地址请求头（与 ThemeMarketController 共享 HeaderMarketBaseURL 常量）。
func (c *ThemeController) marketBaseURL(ctx *gin.Context) string {
	return ctx.GetHeader(HeaderMarketBaseURL)
}

// marketToken 读取官方账号 Token 请求头（与 ThemeMarketController 共享 HeaderMarketToken 常量）。
// 官方代理下载接口要求登录态，安装/更新时必须携带。
func (c *ThemeController) marketToken(ctx *gin.Context) string {
	return ctx.GetHeader(HeaderMarketToken)
}

// List 已安装主题列表 GET /api/v1/themes
func (c *ThemeController) List(ctx *gin.Context) {
	result, err := c.logic.List(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Install 从官方市场安装主题 POST /api/v1/themes/install
func (c *ThemeController) Install(ctx *gin.Context) {
	var r req.ThemeInstallReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, "请提供要安装的官方市场主题 ID", enum.ErrInvalidParam.HttpCode)
		return
	}
	theme, err := c.logic.Install(ctx.Request.Context(), r.ThemeID, r.Version, r.Force, c.marketBaseURL(ctx), c.marketToken(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, theme)
}

// Activate 激活主题实例 POST /api/v1/themes/:id/activate
func (c *ThemeController) Activate(ctx *gin.Context) {
	id := ctx.Param("id")
	theme, err := c.logic.Activate(ctx.Request.Context(), id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, theme)
}

// Uninstall 卸载主题实例 DELETE /api/v1/themes/:id
func (c *ThemeController) Uninstall(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Uninstall(ctx.Request.Context(), id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// Update 从官方市场更新已安装主题到最新版本 POST /api/v1/themes/:id/update
func (c *ThemeController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	theme, err := c.logic.UpdateTheme(ctx.Request.Context(), id, c.marketBaseURL(ctx), c.marketToken(ctx))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, theme)
}
