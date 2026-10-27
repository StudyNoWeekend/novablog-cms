package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"
)

// TestDeriveMixinKey 用官方示例 img/sub key 验证密钥推导
// （期望值经独立实现按同一官方混淆表复核）。
func TestDeriveMixinKey(t *testing.T) {
	raw := "7cd084941338484aae1ad9425b242ee3" + "4932caff0ff746eab6f01bf08b70ac45"
	got := deriveMixinKey(raw)
	want := "ea1db124af337062474693fae24f4ff8"
	if got != want {
		t.Fatalf("deriveMixinKey() = %q, want %q", got, want)
	}
	if len(got) != mixinKeyLen {
		t.Fatalf("mixin key length = %d, want %d", len(got), mixinKeyLen)
	}
}

// TestWBIURLKey 验证从 img/sub URL 中提取密钥本体。
func TestWBIURLKey(t *testing.T) {
	cases := map[string]string{
		"https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b242ee3.png": "7cd084941338484aae1ad9425b242ee3",
		"https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.jpg": "4932caff0ff746eab6f01bf08b70ac45",
		"7cd084941338484aae1ad9425b242ee3":                                  "7cd084941338484aae1ad9425b242ee3",
	}
	for in, want := range cases {
		if got := wbiURLKey(in); got != want {
			t.Fatalf("wbiURLKey(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSignWBIAt 验证签名确定性、参数排序与 w_rid 位置。
func TestSignWBIAt(t *testing.T) {
	key := "ea1db124af3c7062474693fa704f4ff8"

	q := signWBIAt(map[string]string{"bvid": "BV1DTJR6REcK"}, 1700000000, key)
	if !strings.HasPrefix(q, "bvid=BV1DTJR6REcK&wts=1700000000&w_rid=") {
		t.Fatalf("签名 query 排列不符合预期: %s", q)
	}
	wrid := strings.TrimPrefix(q, "bvid=BV1DTJR6REcK&wts=1700000000&w_rid=")
	if len(wrid) != 32 {
		t.Fatalf("w_rid 长度 = %d, want 32: %s", len(wrid), wrid)
	}

	// 同参数同密钥签名必须一致
	if q != signWBIAt(map[string]string{"bvid": "BV1DTJR6REcK"}, 1700000000, key) {
		t.Fatalf("同参数同密钥签名结果不一致")
	}
}

// TestSignWBIAtDeterministicMD5 验证 w_rid 确为 query+密钥 的 MD5。
func TestSignWBIAtDeterministicMD5(t *testing.T) {
	key := "ea1db124af3c7062474693fa704f4ff8"
	q := signWBIAt(map[string]string{"foo": "one"}, 1702204800, key)

	parts := strings.SplitN(q, "&w_rid=", 2)
	sum := md5.Sum([]byte(parts[0] + key))
	want := hex.EncodeToString(sum[:])
	if parts[1] != want {
		t.Fatalf("w_rid = %q, want %q", parts[1], want)
	}
}

// TestFilterWBIChars 验证签名排除字符（!'()*）被移除。
func TestFilterWBIChars(t *testing.T) {
	if got := filterWBIChars("a!b'c(d)e*f"); got != "abcdef" {
		t.Fatalf("filterWBIChars() = %q, want %q", got, "abcdef")
	}
	if got := filterWBIChars("BV1DTJR6REcK"); got != "BV1DTJR6REcK" {
		t.Fatalf("filterWBIChars() = %q", got)
	}
}
