package controller

import (
	"cus-cms/internal/dto/req"
	"cus-cms/internal/logic"
	"cus-cms/utils/response"

	"github.com/gin-gonic/gin"
)

// PortfolioController 摄影作品集控制器结构体。
type PortfolioController struct {
	logic *logic.PortfolioLogic
}

// NewPortfolioController 创建 PortfolioController 实例。
func NewPortfolioController() *PortfolioController {
	return &PortfolioController{logic: logic.NewPortfolioLogic()}
}

// Create 创建作品集 POST /api/v1/portfolios
func (c *PortfolioController) Create(ctx *gin.Context) {
	var r req.CreatePortfolioReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.Create(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetList 获取作品集列表 GET /api/v1/portfolios
func (c *PortfolioController) GetList(ctx *gin.Context) {
	var r req.PortfolioListReq
	if err := ctx.ShouldBindQuery(&r); err != nil {
		response.Error(ctx, "参数错误")
		return
	}
	result, err := c.logic.GetList(ctx, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// GetByID 获取作品集详情 GET /api/v1/portfolios/:id
func (c *PortfolioController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.logic.GetByID(ctx, id)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// Update 更新作品集 PUT /api/v1/portfolios/:id
func (c *PortfolioController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.UpdatePortfolioReq
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

// Delete 删除作品集 DELETE /api/v1/portfolios/:id
func (c *PortfolioController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.logic.Delete(ctx, id); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// AddItem 添加作品项 POST /api/v1/portfolios/:id/items
func (c *PortfolioController) AddItem(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.CreatePortfolioItemReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.AddItem(ctx, id, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// SortItems 批量排序作品项 PUT /api/v1/portfolios/:id/items/sort
func (c *PortfolioController) SortItems(ctx *gin.Context) {
	id := ctx.Param("id")
	var r req.SortPortfolioItemsReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	if err := c.logic.SortItems(ctx, id, &r); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// UpdateItem 更新作品项 PUT /api/v1/portfolios/:id/items/:itemId
func (c *PortfolioController) UpdateItem(ctx *gin.Context) {
	id := ctx.Param("id")
	itemID := ctx.Param("itemId")
	var r req.UpdatePortfolioItemReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "参数错误: "+err.Error())
		return
	}
	result, err := c.logic.UpdateItem(ctx, id, itemID, &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// DeleteItem 删除作品项 DELETE /api/v1/portfolios/:id/items/:itemId
func (c *PortfolioController) DeleteItem(ctx *gin.Context) {
	id := ctx.Param("id")
	itemID := ctx.Param("itemId")
	if err := c.logic.DeleteItem(ctx, id, itemID); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}
