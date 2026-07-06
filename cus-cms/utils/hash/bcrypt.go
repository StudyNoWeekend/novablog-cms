// Package hash 提供密码加密与验证功能。
package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对明文密码进行 bcrypt 哈希加密。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证明文密码与哈希值是否匹配。
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
