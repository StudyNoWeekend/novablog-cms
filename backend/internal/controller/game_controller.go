package controller

import (
	"novablog/enum"
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// GameController 游戏库控制器结构体。
type GameController struct {
	logic *logic.GameLogic
}

// NewGameController 创建 GameController 实例。
func NewGameController() *GameController {
	return &GameController{logic: logic.NewGameLogic()}
}

// Create 创建游戏 POST /api/v1/games
func (c *GameController) Create(ctx *gin.Context) {
	var r req.CreateGameReq
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

// GetList 获取游戏列表 GET /api/v1/games
func (c *GameController) GetList(ctx *gin.Context) {
	var r req.GameListReq
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

// GetByID 获取游戏详情 GET /api/v1/games/:id
func (c *GameController) GetByID(ctx *gin.Context) {
	result, err := c.logic.GetByID(ctx, ctx.Param("id"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Update 更新游戏 PUT /api/v1/games/:id
func (c *GameController) Update(ctx *gin.Context) {
	var r req.UpdateGameReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.Update(ctx, ctx.Param("id"), &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Delete 删除游戏 DELETE /api/v1/games/:id
func (c *GameController) Delete(ctx *gin.Context) {
	if err := c.logic.Delete(ctx, ctx.Param("id")); err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
