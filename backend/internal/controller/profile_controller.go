package controller

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProfileController 个人资料控制器结构体。
type ProfileController struct {
	bloggerLogic *logic.BloggerLogic
	mediaLogic   *logic.MediaLogic
}

// NewProfileController 创建 ProfileController 实例。
func NewProfileController(mediaLogic *logic.MediaLogic) *ProfileController {
	return &ProfileController{
		bloggerLogic: logic.NewBloggerLogic(),
		mediaLogic:   mediaLogic,
	}
}

// GetProfile 获取个人资料 GET /api/v1/profile
func (ctrl *ProfileController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
		return
	}

	result, err := ctrl.bloggerLogic.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, result)
}

// UpdateProfile 更新个人资料 PUT /api/v1/profile
func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
		return
	}

	var updateReq req.UpdateProfileReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		logic.AuthLogger.Warn("更新个人资料请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	result, err := ctrl.bloggerLogic.UpdateProfile(c.Request.Context(), userID.(string), &updateReq)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, result)
}

// UpdateRoles 补选创作方向角色 PUT /api/v1/profile/roles
func (ctrl *ProfileController) UpdateRoles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, enum.ErrUnauthorized.Code, enum.ErrUnauthorized.Msg, enum.ErrUnauthorized.HttpCode)
		return
	}

	var updateReq req.UpdateRolesReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		logic.AuthLogger.Warn("补选创作方向请求参数错误", zap.Error(err))
		response.Fail(c, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	result, err := ctrl.bloggerLogic.UpdateRoles(c.Request.Context(), userID.(string), &updateReq)
	if err != nil {
		var bizErr *enum.BizError
		if errors.As(err, &bizErr) {
			response.Fail(c, bizErr.Code, bizErr.Msg, bizErr.HttpCode)
			return
		}
		response.Fail(c, enum.ErrInternalServer.Code, enum.ErrInternalServer.Msg, enum.ErrInternalServer.HttpCode)
		return
	}

	response.Success(c, result)
}

// UploadIcon 上传博客 Icon 图 POST /api/v1/profile/upload-icon
func (ctrl *ProfileController) UploadIcon(c *gin.Context) {
	url, err := ctrl.uploadImage(c, "icon")
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, gin.H{"url": url})
}

// UploadBackground 上传页面背景图 POST /api/v1/profile/upload-background
func (ctrl *ProfileController) UploadBackground(c *gin.Context) {
	url, err := ctrl.uploadImage(c, "background")
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, gin.H{"url": url})
}

// UploadAvatar 上传头像 POST /api/v1/profile/upload-avatar
func (ctrl *ProfileController) UploadAvatar(c *gin.Context) {
	url, err := ctrl.uploadImage(c, "avatar")
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, gin.H{"url": url})
}

// uploadImage 上传图片的通用方法：校验后统一走媒体库上传链路，
// 登记到 media 表并归入「博主资料」文件夹（头像/图标/背景可在媒体库统一管理）。
// subdir 仅为语义参数（icon、background 或 avatar），物理目录统一归属博主资料文件夹。
func (ctrl *ProfileController) uploadImage(c *gin.Context, subdir string) (string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("请选择上传文件")
	}

	// 限制文件大小 10MB
	if file.Size > 10*1024*1024 {
		return "", fmt.Errorf("文件大小不能超过10MB")
	}

	// 校验文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" && ext != ".svg" && ext != ".ico" {
		return "", fmt.Errorf("不支持的文件格式，仅支持 .jpg/.jpeg/.png/.gif/.webp/.svg/.ico")
	}

	media, err := ctrl.mediaLogic.UploadFile(c.Request.Context(), file, logic.MediaUploadOptions{
		Module: enum.MediaModuleProfile,
	})
	if err != nil {
		return "", err
	}
	return media.URL, nil
}
