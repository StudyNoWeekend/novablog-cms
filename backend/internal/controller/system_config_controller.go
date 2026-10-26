package controller

import (
	"net/http"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/logic"
	"novablog/internal/model"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// SystemConfigController 通用系统配置控制器（站点信息分组 + 配置导出/导入）。
type SystemConfigController struct {
	logic *logic.SystemConfigLogic
}

// NewSystemConfigController 创建 SystemConfigController 实例。
func NewSystemConfigController() *SystemConfigController {
	return &SystemConfigController{logic: logic.NewSystemConfig()}
}

// GetSystemConfig 获取站点信息分组配置 GET /api/v1/system-config
func (c *SystemConfigController) GetSystemConfig(ctx *gin.Context) {
	items, err := c.logic.GetSiteSettings(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	list := make([]res.SystemConfigItemRes, 0, len(items))
	for _, item := range items {
		list = append(list, res.SystemConfigItemRes{
			Key:         item.Key,
			Value:       item.Value,
			Description: item.Description,
		})
	}
	response.Success(ctx, res.SystemConfigListRes{Items: list})
}

// UpdateSystemConfig 更新站点信息分组配置 PUT /api/v1/system-config
// 站点对外地址变化时热更媒体 URL 基址，并在响应中提示检查跨域白名单。
func (c *SystemConfigController) UpdateSystemConfig(ctx *gin.Context) {
	var r req.UpdateSystemConfigReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	siteURLChanged, err := c.logic.UpdateSiteSettings(ctx.Request.Context(), r.Updates)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	// 站点对外地址为媒体 URL 基址的权威值，热更全局变量即时生效
	if siteURLChanged {
		if siteURL := c.logic.GetSiteURL(ctx.Request.Context()); siteURL != "" {
			model.BaseURL = siteURL
		}
	}
	response.Success(ctx, res.SystemConfigUpdateRes{SiteURLChanged: siteURLChanged})
}

// ExportConfig 导出全部运行时配置 GET /api/v1/system-config/export?include_secrets=1
// 默认脱敏：凭据类字段以占位符输出（导入时自动跳过）。
func (c *SystemConfigController) ExportConfig(ctx *gin.Context) {
	includeSecrets := ctx.Query("include_secrets") == "1"
	export, err := c.logic.BuildConfigExport(ctx.Request.Context(), includeSecrets)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	filename := "novablog-config-" + time.Now().Format("20060102-150405") + ".json"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.JSON(http.StatusOK, export)
}

// ImportConfig 导入运行时配置 POST /api/v1/system-config/import
func (c *SystemConfigController) ImportConfig(ctx *gin.Context) {
	var export res.SystemConfigExportRes
	if err := ctx.ShouldBindJSON(&export); err != nil || len(export.Groups) == 0 {
		response.Fail(ctx, enum.ErrInvalidParam.Code, "导入文件格式不正确", enum.ErrInvalidParam.HttpCode)
		return
	}

	applied, err := c.logic.ApplyConfigImport(ctx.Request.Context(), &export)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	// 导入可能包含站点对外地址，热更媒体 URL 基址
	if siteURL := c.logic.GetSiteURL(ctx.Request.Context()); siteURL != "" {
		model.BaseURL = siteURL
	}
	response.Success(ctx, res.SystemConfigImportRes{Applied: applied})
}
