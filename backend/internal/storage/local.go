package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalProvider 本地文件系统存储实现。
type LocalProvider struct {
	uploadDir    string
	pathPrefix   string
	customDomain string
}

// NewLocalProvider 创建本地存储 Provider。
// uploadDir 为本地文件系统目录，会通过 filepath.Clean 规范化。
func NewLocalProvider(uploadDir, pathPrefix, customDomain string) (*LocalProvider, error) {
	return &LocalProvider{
		uploadDir:    filepath.Clean(uploadDir),
		pathPrefix:   pathPrefix,
		customDomain: customDomain,
	}, nil
}

// fullPath 拼接上传目录与对象键，返回完整文件路径。
func (p *LocalProvider) fullPath(key string) string {
	return filepath.Join(p.uploadDir, key)
}

// Upload 上传文件到本地文件系统，返回对象键（相对路径）。
func (p *LocalProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	fullPath := p.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建本地目录失败: %w", err)
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("写入本地文件失败: %w", err)
	}
	return key, nil
}

// Download 下载文件，返回可读流。
func (p *LocalProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	file, err := os.Open(p.fullPath(key))
	if err != nil {
		return nil, fmt.Errorf("打开本地文件失败: %w", err)
	}
	return file, nil
}

// Delete 删除本地文件。
func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	if err := os.Remove(p.fullPath(key)); err != nil {
		return fmt.Errorf("删除本地文件失败: %w", err)
	}
	return nil
}

// Exists 判断本地文件是否存在。
func (p *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := os.Stat(p.fullPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("检查本地文件存在失败: %w", err)
	}
	return true, nil
}

// Type 返回 provider 类型标识。
func (p *LocalProvider) Type() string {
	return "local"
}

// GetThumbURL 本地存储不支持缩略图，返回原始 URL。
func (p *LocalProvider) GetThumbURL(url string, width int) string {
	return url
}
