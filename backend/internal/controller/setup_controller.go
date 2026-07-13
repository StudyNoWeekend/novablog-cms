// Package controller 定义 HTTP 控制器层。
package controller

import (
	"errors"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupController 初始化控制器结构体。
type SetupController struct {
	setupLogic *logic.SetupLogic
}

// NewSetupController 创建 SetupController 实例。
func NewSetupController() *SetupController {
	return &SetupController{
		setupLogic: logic.NewSetupLogic(),
	}
}

// GetStatus 获取系统初始化状态。
func (ctrl *SetupController) GetStatus(c *gin.Context) {
	statusRes, err := ctrl.setupLogic.CheckStatus(c.Request.Context())
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, statusRes)
}

// Init 初始化博主账号。
func (ctrl *SetupController) Init(c *gin.Context) {
	var initReq req.InitReq
	if err := c.ShouldBindJSON(&initReq); err != nil {
		logic.SetupLogger.Warn("初始化请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	initRes, err := ctrl.setupLogic.InitBlogger(c.Request.Context(), &initReq)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, initRes)
}
