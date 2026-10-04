package controller

import (
	"net/http"

	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// ImageSearchController 图片搜索控制器结构体。
type ImageSearchController struct {
	imageSearchLogic *logic.ImageSearchLogic
}

// NewImageSearchController 创建 ImageSearchController 实例。
func NewImageSearchController(imageSearchLogic *logic.ImageSearchLogic) *ImageSearchController {
	return &ImageSearchController{imageSearchLogic: imageSearchLogic}
}

// Search 按类型搜索图片 GET /api/v1/image-search?type=icons|games|books|food&q=
func (ctrl *ImageSearchController) Search(ctx *gin.Context) {
	var r req.ImageSearchReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	list, err := ctrl.imageSearchLogic.SearchImages(ctx.Request.Context(), r.Type, r.Q, r.Limit)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, list)
}

// Save 转存外部图片为本站资源 POST /api/v1/image-search/save
func (ctrl *ImageSearchController) Save(ctx *gin.Context) {
	var r req.SaveImageReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	media, err := ctrl.imageSearchLogic.SaveImage(ctx.Request.Context(), r.URL, r.Module)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, media)
}

// Thumbnail 代理输出外部图片缩略图 GET /api/v1/image-search/thumbnail?url=
// 供搜索弹窗 <img> 引用：浏览器直连外部图床会被反爬/防盗链拦截，统一由后端取图。
func (ctrl *ImageSearchController) Thumbnail(ctx *gin.Context) {
	var r req.ImageThumbReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	data, contentType, err := ctrl.imageSearchLogic.Thumbnail(ctx.Request.Context(), r.URL)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Data(http.StatusOK, contentType, data)
}
