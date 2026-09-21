package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// EquipmentController 个人设备控制器结构体。
type EquipmentController struct {
	logic *logic.EquipmentLogic
}

// NewEquipmentController 创建 EquipmentController 实例。
func NewEquipmentController() *EquipmentController {
	return &EquipmentController{logic: logic.NewEquipmentLogic()}
}

// Create 创建个人设备 POST /api/v1/equipments
func (c *EquipmentController) Create(ctx *gin.Context) {
	var r req.CreateEquipmentReq
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

// GetList 获取个人设备列表 GET /api/v1/equipments
func (c *EquipmentController) GetList(ctx *gin.Context) {
	var r req.EquipmentListReq
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

// GetByID 获取个人设备详情 GET /api/v1/equipments/:id
func (c *EquipmentController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新个人设备 PUT /api/v1/equipments/:id
func (c *EquipmentController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdateEquipmentReq
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

// Delete 删除个人设备 DELETE /api/v1/equipments/:id
func (c *EquipmentController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
