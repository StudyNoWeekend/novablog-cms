// Package factory 提供第三方云盘客户端的统一构建入口。
// 独立成子包以避免与 clouddrive 类型包形成 import cycle。
package factory

import (
	"fmt"

	"novablog/pkg/aliyundrive"
	"novablog/pkg/baidupan"
	"novablog/pkg/clouddrive"
	"novablog/pkg/webdav"
)

// BuildParams 构建客户端所需参数（凭据应为解密后的明文，加解密由 logic 层负责）。
type BuildParams struct {
	Provider string
	// OAuth 盘
	AccessToken string
	DriveID     string
	// WebDAV
	BaseURL  string
	Username string
	Password string

	ProxyURL string
}

// BuildClient 按 provider 构建对应云盘客户端。
func BuildClient(p BuildParams) (clouddrive.Client, error) {
	switch p.Provider {
	case clouddrive.ProviderWebDAV:
		if p.BaseURL == "" {
			return nil, fmt.Errorf("WebDAV 服务器地址不能为空")
		}
		return webdav.New(p.BaseURL, p.Username, p.Password, p.ProxyURL), nil
	case clouddrive.ProviderBaidu:
		return baidupan.New(p.AccessToken, p.ProxyURL), nil
	case clouddrive.ProviderAliyun:
		return aliyundrive.New(p.AccessToken, p.DriveID, p.ProxyURL), nil
	default:
		return nil, fmt.Errorf("不支持的云盘类型: %s", p.Provider)
	}
}
