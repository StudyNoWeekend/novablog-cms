package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MCPKeyController MCP 密钥管理控制器结构体。
type MCPKeyController struct {
	logic *logic.MCPKeyLogic
}

// NewMCPKeyController 创建 MCPKeyController 实例。
func NewMCPKeyController() *MCPKeyController {
	return &MCPKeyController{logic: logic.NewMCPKeyLogic()}
}

// GetList 获取 MCP 密钥列表 GET /api/v1/mcp/keys
func (c *MCPKeyController) GetList(ctx *gin.Context) {
	var r req.MCPKeyListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetList(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Create 创建 MCP 密钥 POST /api/v1/mcp/keys
func (c *MCPKeyController) Create(ctx *gin.Context) {
	var r req.CreateMCPKeyReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新 MCP 密钥 PUT /api/v1/mcp/keys/:id
func (c *MCPKeyController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateMCPKeyReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Update(ctx, id, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除 MCP 密钥 DELETE /api/v1/mcp/keys/:id
func (c *MCPKeyController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
