// Package logic 定义业务逻辑层。
package logic

import (
	"context"
	"errors"
	"time"

	"cus-cms/enum"
	"cus-cms/internal/cache"
	"cus-cms/internal/dto/req"
	"cus-cms/internal/dto/res"
	"cus-cms/internal/model"
	"cus-cms/utils/hash"
	jwtutil "cus-cms/utils/jwt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AccessSecret 访问令牌密钥（由启动时注入）。
var AccessSecret string

// RefreshSecret 刷新令牌密钥（由启动时注入）。
var RefreshSecret string

// AccessExpire 访问令牌过期时间（由启动时注入）。
var AccessExpire time.Duration

// RefreshExpire 刷新令牌过期时间（由启动时注入）。
var RefreshExpire time.Duration

// AuthLogger 认证模块日志记录器。
var AuthLogger *zap.Logger

// AuthLogic 认证业务逻辑结构体。
type AuthLogic struct {
	bloggerModel *model.BloggerModel
	tokenCache   *cache.TokenCache
}

// NewAuthLogic 创建 AuthLogic 实例。
func NewAuthLogic() *AuthLogic {
	return &AuthLogic{
		bloggerModel: model.NewBlogger(),
		tokenCache:   cache.NewTokenCache(),
	}
}

// Login 处理用户登录逻辑：校验用户名密码，生成 Token 对。
func (l *AuthLogic) Login(ctx context.Context, r *req.LoginReq) (*res.LoginRes, error) {
	// 根据用户名查询博主信息
	blogger, err := l.bloggerModel.GetByUsername(ctx, r.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			AuthLogger.Warn("登录失败：用户不存在", zap.String("username", r.Username))
			return nil, enum.ErrLoginFailed
		}
		AuthLogger.Error("登录查询用户失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 校验密码
	if !hash.CheckPassword(r.Password, blogger.PasswordHash) {
		AuthLogger.Warn("登录失败：密码错误", zap.String("username", r.Username))
		return nil, enum.ErrLoginFailed
	}

	// 生成访问令牌
	accessToken, err := jwtutil.GenerateAccessToken(blogger.ID, blogger.Username, AccessSecret, AccessExpire)
	if err != nil {
		AuthLogger.Error("生成访问令牌失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 生成刷新令牌
	refreshToken, err := jwtutil.GenerateRefreshToken(blogger.ID, blogger.Username, RefreshSecret, RefreshExpire)
	if err != nil {
		AuthLogger.Error("生成刷新令牌失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 更新最后登录时间（非关键操作，记录日志但不阻塞登录流程）
	if err := l.bloggerModel.UpdateLastLogin(ctx, blogger.ID); err != nil {
		AuthLogger.Warn("更新最后登录时间失败", zap.Error(err))
	}

	return &res.LoginRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessExpire.Seconds()),
	}, nil
}

// Refresh 使用刷新令牌生成新的 Token 对，并黑名单旧的刷新令牌。
func (l *AuthLogic) Refresh(ctx context.Context, refreshToken string) (*res.LoginRes, error) {
	// 解析刷新令牌
	claims, err := jwtutil.ParseToken(refreshToken, RefreshSecret)
	if err != nil {
		AuthLogger.Warn("刷新令牌解析失败", zap.Error(err))
		return nil, enum.ErrTokenInvalid
	}

	// 检查旧刷新令牌是否已在黑名单中
	blacklisted, err := l.tokenCache.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		AuthLogger.Error("检查黑名单失败", zap.Error(err))
		// 黑名单检查失败不阻塞流程
	}
	if blacklisted {
		AuthLogger.Warn("刷新令牌已在黑名单中", zap.String("jti", claims.ID))
		return nil, enum.ErrTokenInvalid
	}

	// 将旧刷新令牌加入黑名单
	remainingTime := jwtutil.GetTokenRemainingTime(claims)
	if err := l.tokenCache.AddToBlacklist(ctx, claims.ID, remainingTime); err != nil {
		AuthLogger.Error("刷新令牌加入黑名单失败", zap.Error(err))
	}

	// 生成新的访问令牌
	accessToken, err := jwtutil.GenerateAccessToken(claims.UserID, claims.Username, AccessSecret, AccessExpire)
	if err != nil {
		AuthLogger.Error("生成新访问令牌失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	// 生成新的刷新令牌
	newRefreshToken, err := jwtutil.GenerateRefreshToken(claims.UserID, claims.Username, RefreshSecret, RefreshExpire)
	if err != nil {
		AuthLogger.Error("生成新刷新令牌失败", zap.Error(err))
		return nil, enum.ErrInternalServer
	}

	return &res.LoginRes{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(AccessExpire.Seconds()),
	}, nil
}

// Logout 登出：将访问令牌的 JTI 加入黑名单。
func (l *AuthLogic) Logout(ctx context.Context, accessToken string) error {
	claims, err := jwtutil.ParseToken(accessToken, AccessSecret)
	if err != nil {
		AuthLogger.Warn("登出解析令牌失败", zap.Error(err))
		return enum.ErrTokenInvalid
	}

	remainingTime := jwtutil.GetTokenRemainingTime(claims)
	if err := l.tokenCache.AddToBlacklist(ctx, claims.ID, remainingTime); err != nil {
		AuthLogger.Error("令牌加入黑名单失败", zap.Error(err))
		return enum.ErrInternalServer
	}

	return nil
}

// ChangePassword 修改博主密码。
func (l *AuthLogic) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// 查询博主信息
	blogger, err := l.bloggerModel.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			AuthLogger.Warn("修改密码：用户不存在", zap.String("userID", userID))
			return enum.ErrNotFound
		}
		AuthLogger.Error("修改密码查询用户失败", zap.Error(err))
		return enum.ErrInternalServer
	}

	// 验证旧密码
	if !hash.CheckPassword(oldPassword, blogger.PasswordHash) {
		AuthLogger.Warn("修改密码：旧密码错误", zap.String("userID", userID))
		return enum.ErrLoginFailed
	}

	// 哈希新密码
	newHash, err := hash.HashPassword(newPassword)
	if err != nil {
		AuthLogger.Error("新密码哈希失败", zap.Error(err))
		return enum.ErrInternalServer
	}

	blogger.PasswordHash = newHash
	if err := l.bloggerModel.Update(ctx, blogger); err != nil {
		AuthLogger.Error("更新密码失败", zap.Error(err))
		return enum.ErrInternalServer
	}

	return nil
}
