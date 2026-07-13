package controller

import (
	"novablog/internal/dto/req"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// MigrationController 素材迁移控制器结构体。
type MigrationController struct {
	logic *logic.MigrationLogic
}

// NewMigrationController 创建 MigrationController 实例。
func NewMigrationController(logic *logic.MigrationLogic) *MigrationController {
	return &MigrationController{logic: logic}
}

// Analyze 发起迁移分析 POST /storage/migration/analyze
func (c *MigrationController) Analyze(ctx *gin.Context) {
	var r req.MigrationAnalyzeReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	taskID, err := c.logic.Analyze(ctx.Request.Context(), r.TargetProvider)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, gin.H{"task_id": taskID})
}

// GetAnalyzeResult 获取分析结果 GET /storage/migration/analyze/:taskId
func (c *MigrationController) GetAnalyzeResult(ctx *gin.Context) {
	taskID := ctx.Param("taskId")
	result, err := c.logic.GetAnalyzeResult(ctx.Request.Context(), taskID)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, result)
}

// StartMigration 发起迁移执行 POST /storage/migration/start
func (c *MigrationController) StartMigration(ctx *gin.Context) {
	var r req.MigrationStartReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	taskID, err := c.logic.StartMigration(ctx.Request.Context(), &r)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, gin.H{"task_id": taskID})
}

// GetMigrationStatus 获取迁移进度 GET /storage/migration/status/:taskId
func (c *MigrationController) GetMigrationStatus(ctx *gin.Context) {
	taskID := ctx.Param("taskId")
	taskRes, failedItems, err := c.logic.GetMigrationStatus(ctx.Request.Context(), taskID)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, gin.H{
		"task":   taskRes,
		"failed": failedItems,
	})
}
