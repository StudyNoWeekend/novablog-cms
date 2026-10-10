package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"novablog/enum"
	"novablog/internal/logic"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// 上传图片大小上限（20MB）。
const maxImageSize = 20 << 20

// UploadImageIn upload_image 入参。
type UploadImageIn struct {
	Source   string `json:"source" jsonschema:"图片来源：http(s):// 开头的图片 URL，或 base64 编码的图片数据（支持 data URI）"`
	Filename string `json:"filename,omitempty" jsonschema:"文件名（需含扩展名，如 cover.png）；URL 方式默认取 URL 末段"`
}

// UploadImageOut upload_image 出参。
type UploadImageOut struct {
	URL      string `json:"url" jsonschema:"可直接用于封面图/正文配图的 URL"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// uploadImage 上传图片：URL 自动转存，base64 直接落库，统一归入媒体库"文章"文件夹。
func (s *Service) uploadImage(ctx context.Context, _ *sdk.CallToolRequest, in UploadImageIn) (*sdk.CallToolResult, UploadImageOut, error) {
	source := strings.TrimSpace(in.Source)
	if source == "" {
		return nil, UploadImageOut{}, fmt.Errorf("source 不能为空")
	}

	var (
		data        []byte
		filename    string
		contentType string
		err         error
	)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		data, filename, contentType, err = fetchRemoteImage(ctx, source)
	} else {
		data, contentType, err = decodeBase64Image(source)
		filename = "image"
	}
	if err != nil {
		return nil, UploadImageOut{}, err
	}
	if in.Filename != "" {
		filename = sanitizeFilename(in.Filename)
	}
	if path.Ext(filename) == "" {
		filename += extFromMIME(contentType)
	}
	if path.Ext(filename) == "" {
		return nil, UploadImageOut{}, fmt.Errorf("无法识别图片格式，请提供带扩展名的 filename（如 cover.png）")
	}

	media, err := s.mediaLogic.SaveRemoteBytes(ctx, filename, contentType, data, logic.MediaUploadOptions{
		Module: enum.MediaModuleArticle,
	})
	if err != nil {
		return nil, UploadImageOut{}, fmt.Errorf("保存图片失败: %w", err)
	}
	return nil, UploadImageOut{
		URL:      media.URL,
		Filename: media.Filename,
		Size:     media.Size,
		MimeType: media.MimeType,
	}, nil
}

// fetchRemoteImage 抓取远程图片，返回字节、建议文件名与 Content-Type。
func fetchRemoteImage(ctx context.Context, rawURL string) ([]byte, string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, "", "", fmt.Errorf("source 不是合法的 http(s) 图片 URL，也不是 base64 数据")
	}
	if err := assertPublicHost(u.Hostname()); err != nil {
		return nil, "", "", err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", "", fmt.Errorf("构造图片请求失败: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("下载图片失败：HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		return nil, "", "", fmt.Errorf("读取图片内容失败: %w", err)
	}
	if len(data) > maxImageSize {
		return nil, "", "", fmt.Errorf("图片超过 20MB 大小限制")
	}
	if len(data) == 0 {
		return nil, "", "", fmt.Errorf("图片内容为空")
	}

	contentType := detectImageType(data, resp.Header.Get("Content-Type"))
	filename := path.Base(u.Path)
	if filename == "" || filename == "/" || filename == "." {
		filename = "image.png"
	}
	return data, filename, contentType, nil
}

// decodeBase64Image 解码 base64 图片数据（支持 data URI 前缀）。
func decodeBase64Image(source string) ([]byte, string, error) {
	contentType := ""
	if strings.HasPrefix(source, "data:") {
		parts := strings.SplitN(source, ",", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("data URI 格式不正确")
		}
		contentType = strings.TrimPrefix(strings.SplitN(parts[0], ":", 2)[1], "image/")
		source = parts[1]
	}
	data, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return nil, "", fmt.Errorf("source 不是合法的图片 URL 或 base64 数据: %w", err)
	}
	if len(data) > maxImageSize {
		return nil, "", fmt.Errorf("图片超过 20MB 大小限制")
	}
	return data, detectImageType(data, contentType), nil
}

// detectImageType 检测图片 MIME 类型，仅允许常见图片格式。
func detectImageType(data []byte, hint string) string {
	if hint == "" || hint == "application/octet-stream" {
		hint = http.DetectContentType(data)
	}
	hint = strings.ToLower(strings.TrimSpace(strings.SplitN(hint, ";", 2)[0]))
	switch hint {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml", "image/bmp", "image/avif":
		return hint
	default:
		return ""
	}
}

// extFromMIME 将图片 MIME 类型映射为文件扩展名。
func extFromMIME(mt string) string {
	switch mt {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	case "image/avif":
		return ".avif"
	default:
		return ""
	}
}

// sanitizeFilename 清理文件名，仅保留基础名并防止路径穿越。
func sanitizeFilename(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimSpace(name)
	if name == "." || name == "/" || name == "" {
		return ""
	}
	return name
}

// assertPublicHost 阻止抓取私网/回环地址（防止 MCP 被引导访问内网资源）。
func assertPublicHost(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("图片 URL 缺少主机名")
	}
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("解析图片主机失败: %w", err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("不允许抓取内网地址的图片")
		}
	}
	return nil
}
