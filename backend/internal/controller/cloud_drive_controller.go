package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// CloudDriveController 云盘导入控制器。
type CloudDriveController struct {
	logic       *logic.CloudDriveLogic
	importLogic *logic.CloudImportLogic
}

// NewCloudDriveController 创建 CloudDriveController 实例。
func NewCloudDriveController(logic *logic.CloudDriveLogic, importLogic *logic.CloudImportLogic) *CloudDriveController {
	return &CloudDriveController{logic: logic, importLogic: importLogic}
}

// GetConfigs 获取云盘应用凭据列表 GET /clouddrive/config
func (c *CloudDriveController) GetConfigs(ctx *gin.Context) {
	configs, err := c.logic.ListConfigs(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, configs)
}

// UpsertConfig 保存云盘应用凭据 POST /clouddrive/config
func (c *CloudDriveController) UpsertConfig(ctx *gin.Context) {
	var r req.CloudDriveConfigReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	if err := c.logic.UpsertConfig(ctx.Request.Context(), &r); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// DeleteConfig 删除云盘应用凭据 DELETE /clouddrive/config/:provider
func (c *CloudDriveController) DeleteConfig(ctx *gin.Context) {
	if err := c.logic.DeleteConfig(ctx.Request.Context(), ctx.Param("provider")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// GetAuthorizeURL 生成 OAuth 授权地址 POST /clouddrive/authorize/:provider
func (c *CloudDriveController) GetAuthorizeURL(ctx *gin.Context) {
	resData, err := c.logic.GetAuthorizeURL(ctx.Request.Context(), ctx.Param("provider"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, resData)
}

// OAuthCallback OAuth 授权回调（公开路由）：处理后渲染自关闭提示页。
func (c *CloudDriveController) OAuthCallback(ctx *gin.Context) {
	provider := ctx.Param("provider")
	code := ctx.Query("code")
	state := ctx.Query("state")

	result, err := c.logic.HandleCallback(ctx.Request.Context(), provider, code, state)
	if err != nil {
		result = &logic.OAuthCallbackResult{Success: false, Message: err.Error()}
	}
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, oauthCallbackHTML(result))
}

// oauthCallbackHTML 渲染回调提示页：向 opener 通知结果并尝试自动关闭窗口。
func oauthCallbackHTML(result *logic.OAuthCallbackResult) string {
	payload := map[string]interface{}{
		"success": result.Success,
		"message": result.Message,
		"account": result.Account,
	}
	raw := oauthPayloadJSON(payload)
	status := "ok"
	if !result.Success {
		status = "error"
	}
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>云盘授权</title></head>
<body style="font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#f5f6fa;">
<div style="text-align:center;">
<p style="font-size:20px;margin-bottom:8px;">` + result.Message + `</p>
<p style="color:#8A93A8;font-size:14px;">请返回管理后台查看账号状态</p>
</div>
<script>
try { if (window.opener) { window.opener.postMessage({ source: 'novablog-clouddrive', status: '` + status + `', payload: ` + raw + ` }, '*'); } } catch (e) {}
setTimeout(function () { try { window.close(); } catch (e) {} }, 1500);
</script>
</body>
</html>`
}

// oauthPayloadJSON 序列化回调载荷（HTML 内嵌）。
func oauthPayloadJSON(payload map[string]interface{}) string {
	b, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// ListAccounts 获取云盘账号列表 GET /clouddrive/accounts
func (c *CloudDriveController) ListAccounts(ctx *gin.Context) {
	accounts, err := c.logic.ListAccounts(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, accounts)
}

// AddWebDAVAccount 添加 WebDAV 账号 POST /clouddrive/accounts/webdav
func (c *CloudDriveController) AddWebDAVAccount(ctx *gin.Context) {
	var r req.WebDAVAccountReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	account, err := c.logic.AddWebDAVAccount(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, account)
}

// TestWebDAV 测试 WebDAV 连通性 POST /clouddrive/accounts/webdav/test
func (c *CloudDriveController) TestWebDAV(ctx *gin.Context) {
	var r req.WebDAVTestReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	if err := c.logic.TestWebDAV(ctx.Request.Context(), &r); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// DeleteAccount 解绑云盘账号 DELETE /clouddrive/accounts/:id
func (c *CloudDriveController) DeleteAccount(ctx *gin.Context) {
	if err := c.logic.DeleteAccount(ctx.Request.Context(), ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ListFiles 浏览云盘目录 GET /clouddrive/accounts/:id/files
func (c *CloudDriveController) ListFiles(ctx *gin.Context) {
	limit := 50
	if v := ctx.Query("limit"); v != "" {
		if n := parseIntParam(v); n > 0 {
			limit = n
		}
	}
	result, err := c.logic.ListFiles(ctx.Request.Context(), ctx.Param("id"), ctx.Query("parent"), ctx.Query("cursor"), limit)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// StartImport 发起导入 POST /clouddrive/import/start
func (c *CloudDriveController) StartImport(ctx *gin.Context) {
	var r req.CloudImportStartReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	taskID, err := c.importLogic.StartImport(ctx.Request.Context(), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"task_id": taskID})
}

// GetImportStatus 查询导入进度 GET /clouddrive/import/status/:taskId
func (c *CloudDriveController) GetImportStatus(ctx *gin.Context) {
	result, err := c.importLogic.GetImportStatus(ctx.Request.Context(), ctx.Param("taskId"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetImportTasks 查询导入任务历史 GET /clouddrive/import/tasks
func (c *CloudDriveController) GetImportTasks(ctx *gin.Context) {
	page := parseIntParam(ctx.DefaultQuery("page", "1"))
	pageSize := parseIntParam(ctx.DefaultQuery("page_size", "20"))
	tasks, total, err := c.importLogic.GetImportTasks(ctx.Request.Context(), page, pageSize)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{
		"list":         tasks,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// parseIntParam 安全解析整数参数。
func parseIntParam(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
