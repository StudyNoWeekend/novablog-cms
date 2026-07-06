// Package logic 定义业务逻辑层。
package logic

import (
	"context"

	"cus-cms/enum"
	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"
	"cus-cms/utils/hash"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SetupLogger 初始化模块日志记录器。
var SetupLogger *zap.Logger

// SetupLogic 初始化业务逻辑结构体。
type SetupLogic struct {
	bloggerModel *model.BloggerModel
}

// NewSetupLogic 创建 SetupLogic 实例。
func NewSetupLogic() *SetupLogic {
	return &SetupLogic{
		bloggerModel: model.NewBlogger(),
	}
}

// CheckStatus 检查系统初始化状态。
func (l *SetupLogic) CheckStatus(ctx context.Context) (*res.StatusRes, error) {
	count, err := l.bloggerModel.Count(ctx)
	if err != nil {
		SetupLogger.Error("查询博主数量失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	return &res.StatusRes{
		Initialized: count > 0,
	}, nil
}

// InitBlogger 初始化博主账号。
func (l *SetupLogic) InitBlogger(ctx context.Context, r *req.InitReq) (*res.InitRes, error) {
	// 检查是否已初始化
	count, err := l.bloggerModel.Count(ctx)
	if err != nil {
		SetupLogger.Error("查询博主数量失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}
	if count > 0 {
		SetupLogger.Warn("系统已初始化，无法重复创建")
		return nil, enum.ErrAlreadyInitialized
	}

	// 哈希密码
	passwordHash, err := hash.HashPassword(r.Password)
	if err != nil {
		SetupLogger.Error("密码哈希失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 设置昵称，如果未提供则使用用户名
	nickname := r.Nickname
	if nickname == "" {
		nickname = r.Username
	}

	// 创建博主记录
	blogger := &model.Blogger{
		ID:           uuid.New().String(),
		Username:     r.Username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	}

	if err := l.bloggerModel.Create(ctx, blogger); err != nil {
		SetupLogger.Error("创建博主失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	SetupLogger.Info("博主账号初始化成功", zap.String("username", r.Username))

	return &res.InitRes{
		Success: true,
		Message: "初始化成功",
	}, nil
}