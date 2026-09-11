// Package jwt 提供 JWT Token 的生成与解析功能。
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims 自定义 JWT 声明，包含用户信息。
type JWTClaims struct {
	UserID   string `json:"user_id"`  // 用户 ID
	Username string `json:"username"` // 用户名
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌。
func GenerateAccessToken(userID, username, secret string, expire time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken 生成刷新令牌（过期时间较长）。
func GenerateRefreshToken(userID, username, secret string, expire time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证 JWT Token，返回声明信息。
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// GetTokenRemainingTime 获取 Token 剩余有效时间。
func GetTokenRemainingTime(claims *JWTClaims) time.Duration {
	if claims == nil || claims.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}
	return remaining
}
