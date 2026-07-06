package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioProvider MinIO 存储实现。
type MinioProvider struct {
	client       *minio.Client
	bucketName   string
	endpoint     string
	useSSL       bool
	pathPrefix   string
	customDomain string
}

// NewMinioProvider 创建 MinIO Provider。
// endpoint 不含 scheme，如 play.min.io 或 localhost:9000。
func NewMinioProvider(endpoint, region, bucket, accessKey, accessSecret, pathPrefix, customDomain string, useSSL bool) (*MinioProvider, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, accessSecret, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO client 失败: %w", err)
	}

	return &MinioProvider{
		client:       client,
		bucketName:   bucket,
		endpoint:     endpoint,
		useSSL:       useSSL,
		pathPrefix:   pathPrefix,
		customDomain: customDomain,
	}, nil
}

// buildKey 在 key 前拼接 pathPrefix 前缀。
func (p *MinioProvider) buildKey(key string) string {
	if p.pathPrefix == "" {
		return key
	}
	return path.Join(strings.Trim(p.pathPrefix, "/"), key)
}

// buildURL 构造对象可访问的 URL。
func (p *MinioProvider) buildURL(key string) string {
	if p.customDomain != "" {
		return strings.TrimRight(p.customDomain, "/") + "/" + key
	}
	scheme := "http"
	if p.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, p.endpoint, p.bucketName, key)
}

// Upload 上传文件并返回可访问 URL。
func (p *MinioProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	objectKey := p.buildKey(key)
	if _, err := p.client.PutObject(ctx, p.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return "", fmt.Errorf("MinIO 上传失败: %w", err)
	}
	return p.buildURL(objectKey), nil
}

// Download 下载文件，返回可读流。
func (p *MinioProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey := p.buildKey(key)
	obj, err := p.client.GetObject(ctx, p.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("MinIO 下载失败: %w", err)
	}
	return obj, nil
}

// Delete 删除对象。
func (p *MinioProvider) Delete(ctx context.Context, key string) error {
	objectKey := p.buildKey(key)
	if err := p.client.RemoveObject(ctx, p.bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("MinIO 删除失败: %w", err)
	}
	return nil
}

// Exists 判断对象是否存在。
func (p *MinioProvider) Exists(ctx context.Context, key string) (bool, error) {
	objectKey := p.buildKey(key)
	_, err := p.client.StatObject(ctx, p.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		// 404 表示对象不存在，连接正常
		if errResp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("MinIO 检查对象存在失败: %w", err)
	}
	return true, nil
}

// Type 返回 provider 类型标识。
func (p *MinioProvider) Type() string {
	return "minio"
}
