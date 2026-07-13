package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunProvider 阿里云 OSS 存储实现。
type AliyunProvider struct {
	client       *oss.Client
	bucket       *oss.Bucket
	endpoint     string
	bucketName   string
	pathPrefix   string
	customDomain string
}

// NewAliyunProvider 创建阿里云 OSS Provider。
// endpoint 应为含 https 的 region 级端点，如 https://oss-cn-hangzhou.aliyuncs.com。
// 使用 V4 签名，region 为 V4 签名必填项。
func NewAliyunProvider(endpoint, region, bucket, accessKey, accessSecret, pathPrefix, customDomain string) (*AliyunProvider, error) {
	client, err := oss.New(endpoint, accessKey, accessSecret,
		oss.AuthVersion(oss.AuthV4), oss.Region(region))
	if err != nil {
		return nil, fmt.Errorf("创建阿里云 OSS client 失败: %w", err)
	}

	bkt, err := client.Bucket(bucket)
	if err != nil {
		return nil, fmt.Errorf("获取阿里云 OSS bucket 失败: %w", err)
	}

	return &AliyunProvider{
		client:       client,
		bucket:       bkt,
		endpoint:     endpoint,
		bucketName:   bucket,
		pathPrefix:   pathPrefix,
		customDomain: customDomain,
	}, nil
}

// buildKey 在 key 前拼接 pathPrefix 前缀。
func (p *AliyunProvider) buildKey(key string) string {
	if p.pathPrefix == "" {
		return key
	}
	return path.Join(strings.Trim(p.pathPrefix, "/"), key)
}

// buildURL 构造对象可访问的 URL。
func (p *AliyunProvider) buildURL(key string) string {
	if p.customDomain != "" {
		return strings.TrimRight(p.customDomain, "/") + "/" + key
	}
	// endpoint 形如 https://oss-cn-hangzhou.aliyuncs.com，去掉 scheme
	host := strings.TrimPrefix(p.endpoint, "https://")
	host = strings.TrimPrefix(host, "http://")
	return fmt.Sprintf("https://%s.%s/%s", p.bucketName, host, key)
}

// Upload 上传文件并返回可访问 URL。
func (p *AliyunProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	objectKey := p.buildKey(key)
	var opts []oss.Option
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	if err := p.bucket.PutObject(objectKey, reader, opts...); err != nil {
		return "", fmt.Errorf("阿里云 OSS 上传失败: %w", err)
	}
	return p.buildURL(objectKey), nil
}

// Download 下载文件，返回可读流。
func (p *AliyunProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey := p.buildKey(key)
	body, err := p.bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("阿里云 OSS 下载失败: %w", err)
	}
	return body, nil
}

// Delete 删除对象。
func (p *AliyunProvider) Delete(ctx context.Context, key string) error {
	objectKey := p.buildKey(key)
	if err := p.bucket.DeleteObject(objectKey); err != nil {
		return fmt.Errorf("阿里云 OSS 删除失败: %w", err)
	}
	return nil
}

// Exists 判断对象是否存在。
func (p *AliyunProvider) Exists(ctx context.Context, key string) (bool, error) {
	objectKey := p.buildKey(key)
	exist, err := p.bucket.IsObjectExist(objectKey)
	if err != nil {
		return false, fmt.Errorf("阿里云 OSS 检查对象存在失败: %w", err)
	}
	return exist, nil
}

// Type 返回 provider 类型标识。
func (p *AliyunProvider) Type() string {
	return "aliyun"
}

// GetThumbURL 阿里云 OSS 暂不支持缩略图 URL，返回原始 URL。
func (p *AliyunProvider) GetThumbURL(url string, width int) string {
	return url
}
