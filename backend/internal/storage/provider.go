// Package storage 提供对象存储多平台抽象层。
package storage

import (
	"context"
	"io"
)

// StorageProvider 对象存储抽象接口。
type StorageProvider interface {
	// Upload 上传文件，key 为对象键，返回可访问的 URL。
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (url string, err error)
	// Download 下载文件，返回可读流。
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除对象。
	Delete(ctx context.Context, key string) error
	// Exists 判断对象是否存在。
	Exists(ctx context.Context, key string) (bool, error)
	// Type 返回 provider 类型标识。
	Type() string
	// GetThumbURL 根据原始 URL 生成指定宽度的缩略图 URL。
	// 不支持缩略图的 Provider 返回原始 URL。
	GetThumbURL(url string, width int) string
}
