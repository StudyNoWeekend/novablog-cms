// Package crypto 提供对称加密与解密功能。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt 使用 AES-GCM 加密明文，返回 base64 编码的密文。
// key 通过 SHA-256 派生为 32 字节的 AES-256 密钥。
// 密文格式：base64(nonce + ciphertext)
func Encrypt(plaintext, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 派生 32 字节 AES-256 密钥
	aesKey := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal 将 nonce 拼接到密文前，最终结果为 nonce + ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的 AES-GCM 密文。
// 对 base64 解码后，前 12 字节为 nonce，其余为密文。
func Decrypt(ciphertext, key string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 派生 32 字节 AES-256 密钥
	aesKey := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
