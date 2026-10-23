// Package clouddrive 定义第三方云盘的统一客户端接口与公共类型。
// 客户端均为无状态实现：token/账密由调用方构造时传入，token 生命周期由 logic 层管理。
package clouddrive

import (
	"context"
	"io"
)

// File 云盘文件信息。
type File struct {
	FileID   string // 盘方文件 ID（WebDAV 为完整路径）
	ParentID string // 父目录标识（WebDAV 为父路径）
	Path     string // 盘内完整路径（WebDAV 与 FileID 相同）
	Name     string // 文件/文件夹名
	IsDir    bool
	Size     int64
	Ext      string // 小写扩展名（含点），文件夹为空
}

// UserInfo 云盘账号信息。
type UserInfo struct {
	OpenUserID  string
	DisplayName string
	QuotaTotal  int64
	QuotaUsed   int64
	DriveID     string // 阿里云盘 default_drive_id；其他盘为空
}

// TokenResult OAuth 换取/刷新凭据结果。
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // 秒
}

// Client 第三方云盘客户端统一接口。
type Client interface {
	// List 列出指定目录内容。parent 为目录标识（WebDAV=路径，百度=盘内路径，
	// 阿里云盘=file_id，根目录传空）；cursor 为翻页游标（百度为页码，首页传空）。
	// 返回文件列表与下一页游标（空串表示没有更多数据）。
	List(ctx context.Context, parent string, cursor string, limit int) ([]File, string, error)
	// Download 获取文件下载流。内部处理直链获取、UA/鉴权头，返回可读流；
	// size 为实际字节数（未知时返回 0），contentType 为响应 Content-Type。
	Download(ctx context.Context, fileID string) (reader io.ReadCloser, size int64, contentType string, err error)
	// UserInfo 获取账号信息（昵称、容量），WebDAV 仅校验连通性。
	UserInfo(ctx context.Context) (UserInfo, error)
}

// Provider 常量。
const (
	ProviderWebDAV = "webdav"
	ProviderBaidu  = "baidu"
	ProviderAliyun = "aliyun"
)

// Supported 判断 provider 是否受支持。
func Supported(provider string) bool {
	switch provider {
	case ProviderWebDAV, ProviderBaidu, ProviderAliyun:
		return true
	}
	return false
}
