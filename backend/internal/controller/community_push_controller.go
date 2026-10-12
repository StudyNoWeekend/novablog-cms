package controller

import (
	"strconv"

	"novablog/enum"
	"novablog/internal/logic"
	"novablog/utils/response"

	"github.com/gin-gonic/gin"
)

// CommunityPushController CMS 内容推送到 NovaBlog 社区的管理控制器。
type CommunityPushController struct {
	logic *logic.CommunityPushLogic
}

// NewCommunityPushController 创建 CommunityPushController 实例。
func NewCommunityPushController() *CommunityPushController {
	return &CommunityPushController{logic: logic.NewCommunityPushLogic()}
}

// GetConfig 获取推送配置 GET /api/v1/community-push/config
func (c *CommunityPushController) GetConfig(ctx *gin.Context) {
	result, err := c.logic.GetPushConfig(ctx)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// UpdateConfig 保存推送配置 PUT /api/v1/community-push/config
func (c *CommunityPushController) UpdateConfig(ctx *gin.Context) {
	var r logic.UpdatePushConfigRequest
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.UpdatePushConfig(ctx, &r)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// TestConnection 连通性自检 POST /api/v1/community-push/test
func (c *CommunityPushController) TestConnection(ctx *gin.Context) {
	info, err := c.logic.TestConnection(ctx)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, info)
}

// GetState 单个内容的推送状态 GET /api/v1/community-push/state?target_type=&target_id=
func (c *CommunityPushController) GetState(ctx *gin.Context) {
	targetType := ctx.Query("target_type")
	targetID := ctx.Query("target_id")
	if targetType == "" || targetID == "" {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	result, err := c.logic.GetState(ctx, targetType, targetID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetStates 批量获取推送状态 GET /api/v1/community-push/states?target_type=&ids=a,b,c
func (c *CommunityPushController) GetStates(ctx *gin.Context) {
	targetType := ctx.Query("target_type")
	idsParam := ctx.Query("ids")
	if targetType == "" || idsParam == "" {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}
	ids := splitIDsParam(idsParam)
	result, err := c.logic.GetStates(ctx, targetType, ids)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// Sync 推送/下架/删除 POST /api/v1/community-push/sync
func (c *CommunityPushController) Sync(ctx *gin.Context) {
	var r struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   string `json:"target_id" binding:"required"`
		Action     string `json:"action" binding:"required,oneof=push hide remove"`
	}
	if err := ctx.ShouldBindJSON(&r); err != nil {
		response.Fail(ctx, enum.ErrInvalidParam.Code, enum.ErrInvalidParam.Msg, enum.ErrInvalidParam.HttpCode)
		return
	}

	var (
		result *logic.SyncActionResult
		err    error
	)
	switch r.Action {
	case "push":
		result, err = c.logic.PushToCommunity(ctx, r.TargetType, r.TargetID)
	case "hide":
		result, err = c.logic.HideFromCommunity(ctx, r.TargetType, r.TargetID)
	case "remove":
		result, err = c.logic.RemoveFromCommunity(ctx, r.TargetType, r.TargetID)
	}
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// GetRecords 同步记录分页 GET /api/v1/community-push/records
func (c *CommunityPushController) GetRecords(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))
	result, total, err := c.logic.ListRecords(ctx, page, size, ctx.Query("target_type"), ctx.Query("sync_status"))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"total": total, "page": page, "size": size, "items": result})
}

// GetTypes 代理社区内容类型字典 GET /api/v1/community-push/types
func (c *CommunityPushController) GetTypes(ctx *gin.Context) {
	result, err := c.logic.GetCommunityTypes(ctx)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.Success(ctx, result)
}

// splitIDsParam 逗号分隔 ID 串切分。
func splitIDsParam(raw string) []string {
	ids := make([]string, 0, 8)
	current := make([]rune, 0, 36)
	for _, r := range raw {
		if r == ',' {
			if len(current) > 0 {
				ids = append(ids, string(current))
				current = current[:0]
			}
			continue
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		ids = append(ids, string(current))
	}
	return ids
}
