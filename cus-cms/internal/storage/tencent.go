package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentProvider 腾讯云 COS 存储实现。
type TencentProvider struct {
	client       *cos.Client
	bucketURL    string
	pathPrefix   string
	customDomain string
}

// NewTencentProvider 创建腾讯云 COS Provider。
// bucketURL 是完整桶 URL，如 https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com。
// 如果 bucketURL 没有 scheme（不以 http:// 或 https:// 开头），自动补 https:// 前缀。
// region 非空时设置 ServiceURL 为 https://cos.{region}.myqcloud.com。
func NewTencentProvider(bucketURL, secretID, secretKey, region, pathPrefix, customDomain string) (*TencentProvider, error) {
	if !strings.HasPrefix(bucketURL, "http://") && !strings.HasPrefix(bucketURL, "https://") {
		bucketURL = "https://" + bucketURL
	}
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析腾讯云 COS bucket URL 失败: %w", err)
	}

	baseURL := &cos.BaseURL{BucketURL: u}
	if region != "" {
		serviceURL, err := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", region))
		if err != nil {
			return nil, fmt.Errorf("解析腾讯云 COS service URL 失败: %w", err)
		}
		baseURL.ServiceURL = serviceURL
	}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return &TencentProvider{
		client:       client,
		bucketURL:    bucketURL,
		pathPrefix:   pathPrefix,
		customDomain: customDomain,
	}, nil
}

// buildKey 在 key 前拼接 pathPrefix 前缀。
func (p *TencentProvider) buildKey(key string) string {
	if p.pathPrefix == "" {
		return key
	}
	return path.Join(strings.Trim(p.pathPrefix, "/"), key)
}

// buildURL 构造对象可访问的 URL。
func (p *TencentProvider) buildURL(key string) string {
	if p.customDomain != "" {
		return strings.TrimRight(p.customDomain, "/") + "/" + key
	}
	return strings.TrimRight(p.bucketURL, "/") + "/" + key
}

// Upload 上传文件并返回可访问 URL。
func (p *TencentProvider) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	objectKey := p.buildKey(key)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	if _, err := p.client.Object.Put(ctx, objectKey, reader, opt); err != nil {
		return "", fmt.Errorf("腾讯云 COS 上传失败: %w", err)
	}
	return p.buildURL(objectKey), nil
}

// Download 下载文件，返回可读流。
func (p *TencentProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey := p.buildKey(key)
	resp, err := p.client.Object.Get(ctx, objectKey, nil)
	if err != nil {
		return nil, fmt.Errorf("腾讯云 COS 下载失败: %w", err)
	}
	return resp.Body, nil
}

// Delete 删除对象。
func (p *TencentProvider) Delete(ctx context.Context, key string) error {
	objectKey := p.buildKey(key)
	if _, err := p.client.Object.Delete(ctx, objectKey); err != nil {
		return fmt.Errorf("腾讯云 COS 删除失败: %w", err)
	}
	return nil
}

// Exists 判断对象是否存在。
func (p *TencentProvider) Exists(ctx context.Context, key string) (bool, error) {
	objectKey := p.buildKey(key)
	exist, err := p.client.Object.IsExist(ctx, objectKey)
	if err != nil {
		return false, fmt.Errorf("腾讯云 COS 检查对象存在失败: %w", err)
	}
	return exist, nil
}

// Type 返回 provider 类型标识。
func (p *TencentProvider) Type() string {
	return "tencent"
}
