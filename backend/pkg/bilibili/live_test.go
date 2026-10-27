//go:build live

// 真实联网验证测试（默认不随 CI 运行）：
//
//	go test ./pkg/bilibili/ -tags live -run TestFetchVideoInfoLive -v
package bilibili

import (
	"context"
	"testing"
)

func TestFetchVideoInfoLive(t *testing.T) {
	info, err := FetchVideoInfo(context.Background(), "BV1DTJR6REcK")
	if err != nil {
		t.Fatalf("FetchVideoInfo() error = %v", err)
	}
	t.Logf("title=%q pic=%q owner=%q cid=%d duration=%d pages=%d",
		info.Title, info.Pic, info.OwnerName, info.CID, info.Duration, len(info.Pages))
	if info.Title == "" || info.CID == 0 {
		t.Fatalf("解析结果不完整: %+v", info)
	}
}
