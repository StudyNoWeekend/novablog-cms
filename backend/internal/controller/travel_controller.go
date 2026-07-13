package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// TravelGuideController 旅行攻略控制器结构体。
type TravelGuideController struct {
	logic *logic.TravelGuideLogic
}

// NewTravelGuideController 创建 TravelGuideController 实例。
func NewTravelGuideController() *TravelGuideController {
	return &TravelGuideController{logic: logic.NewTravelGuideLogic()}
}

// Create 创建旅行攻略 POST /api/v1/travels
func (c *TravelGuideController) Create(ctx *gin.Context) {
	var r req.CreateTravelGuideReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	if r.Status == 0 {
		r.Status = 1 // 默认草稿
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetList 获取旅行攻略列表 GET /api/v1/travels
func (c *TravelGuideController) GetList(ctx *gin.Context) {
	var r req.TravelGuideListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.GetList(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetDetail 获取旅行攻略详情 GET /api/v1/travels/:id
func (c *TravelGuideController) GetDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetDetail(ctx, id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Update 更新旅行攻略 PUT /api/v1/travels/:id
func (c *TravelGuideController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateTravelGuideReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.Update(ctx, id, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Delete 删除旅行攻略 DELETE /api/v1/travels/:id
func (c *TravelGuideController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// UpdateStatus 更新旅行攻略状态 PUT /api/v1/travels/:id/status
func (c *TravelGuideController) UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateTravelGuideStatusReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: status 必须为 1/2/3")
		return
	}
	if err := c.logic.UpdateStatus(ctx, id, &r); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}
