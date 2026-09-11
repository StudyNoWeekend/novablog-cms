package controller

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/internal/storage"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProfileController 个人资料控制器结构体。
type ProfileController struct {
	bloggerLogic *logic.BloggerLogic
	manager      *storage.Manager
}

// NewProfileController 创建 ProfileController 实例。
func NewProfileController(manager *storage.Manager) *ProfileController {
	return &ProfileController{
		bloggerLogic: logic.NewBloggerLogic(),
		manager:      manager,
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

// uploadImage 上传图片到对象存储的通用方法。
// subdir 为子目录（icon、background 或 avatar）。
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

	provider := ctrl.manager.GetProviderOrReload(c.Request.Context())
	if provider == nil {
		return "", fmt.Errorf("对象存储未配置，请在存储配置页面创建并激活存储配置")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 生成对象 key：profile/{subdir}/{date}/{uuid}.ext
	now := time.Now()
	dateDir := now.Format("2006/01")
	storeFilename := uuid.New().String() + ext
	key := fmt.Sprintf("profile/%s/%s/%s", subdir, dateDir, storeFilename)

	if activeCfg := ctrl.manager.GetActiveConfig(); activeCfg != nil && activeCfg.PathPrefix != "" {
		key = fmt.Sprintf("%s/%s", strings.Trim(activeCfg.PathPrefix, "/"), key)
	}

	mimeType := getProfileMimeType(ext)
	url, err := provider.Upload(c.Request.Context(), key, src, file.Size, mimeType)
	if err != nil {
		return "", fmt.Errorf("上传文件到对象存储失败: %w", err)
	}

	return url, nil
}

// getProfileMimeType 根据扩展名返回 MIME 类型。
func getProfileMimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}
