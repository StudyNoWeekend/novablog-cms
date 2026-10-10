package logic

import (
	"strings"
	"testing"
)

// TestGenerateKey 验证密钥生成格式与随机性。
func TestGenerateKey(t *testing.T) {
	key, prefix, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	// 前缀 + 48 位 hex
	if !strings.HasPrefix(key, MCPKeyPrefix) {
		t.Errorf("key 应以 %s 开头，实际 %s", MCPKeyPrefix, key)
	}
	if len(key) != len(MCPKeyPrefix)+48 {
		t.Errorf("key 长度应为 %d，实际 %d", len(MCPKeyPrefix)+48, len(key))
	}
	// 展示前缀 = 前缀 + 前 8 位随机字符
	wantPrefix := key[:len(MCPKeyPrefix)+8]
	if prefix != wantPrefix {
		t.Errorf("prefix 应为 %s，实际 %s", wantPrefix, prefix)
	}

	// 两次生成不应重复
	key2, _, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() second error = %v", err)
	}
	if key == key2 {
		t.Error("两次生成的密钥不应相同")
	}
}

// TestHashKey 验证哈希稳定性与格式。
func TestHashKey(t *testing.T) {
	key := "nbt_mcp_0123456789abcdef0123456789abcdef0123456789abcdef"
	h1 := HashKey(key)
	h2 := HashKey(key)
	if h1 != h2 {
		t.Error("同一密钥的哈希应稳定")
	}
	if len(h1) != 64 {
		t.Errorf("SHA-256 十六进制哈希长度应为 64，实际 %d", len(h1))
	}
	if HashKey(key+"x") == h1 {
		t.Error("不同密钥的哈希不应相同")
	}
}
