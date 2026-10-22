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
	"novablog/internal/dto/req"
	"novablog/internal/dto/res"
	"novablog/internal/logic"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// 上传图片大小上限（20MB）。
const maxImageSize = 20 << 20

// UploadImageIn upload_image 入参。
type UploadImageIn struct {
	Source   string `json:"source" jsonschema:"图片来源：http(s):// 开头的图片 URL，或 base64 编码的图片数据（支持 data URI）"`
	Filename string `json:"filename,omitempty" jsonschema:"文件名（需含扩展名，如 cover.png）；URL 方式默认取 URL 末段"`
	Module   string `json:"module,omitempty" jsonschema:"归入的媒体库模块文件夹，可选：article(默认)/recipe/book/game/tech_stack/travel/equipment/project/video/music/portfolio/profile"`
}

// UploadImageOut upload_image 出参。
type UploadImageOut struct {
	URL      string `json:"url" jsonschema:"可直接用于封面图/正文配图的 URL"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// uploadImage 上传图片：URL 自动转存，base64 直接落库，按 module 参数归入对应模块文件夹（默认 article）。
func (s *Service) uploadImage(ctx context.Context, _ *sdk.CallToolRequest, in UploadImageIn) (*sdk.CallToolResult, UploadImageOut, error) {
	source := strings.TrimSpace(in.Source)
	if source == "" {
		return nil, UploadImageOut{}, fmt.Errorf("source 不能为空")
	}
	module := strings.TrimSpace(in.Module)
	if module == "" {
		module = enum.MediaModuleArticle
	}
	if _, ok := enum.MediaModuleFolders[module]; !ok {
		return nil, UploadImageOut{}, fmt.Errorf("非法的媒体模块 %q，可选：%s", module, strings.Join(enum.MediaModuleOrder, "/"))
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
		Module: module,
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

// ListMediaIn list_media 入参。
type ListMediaIn struct {
	FileType *int16 `json:"file_type,omitempty" jsonschema:"按类型筛选: 0=其他 1=图片 2=视频 3=音频"`
	Keyword  string `json:"keyword,omitempty" jsonschema:"按文件名模糊搜索"`
	FolderID string `json:"folder_id,omitempty" jsonschema:"按文件夹筛选：root 表示根目录，其他值为文件夹 ID；不传返回全部媒体"`
	Page     int    `json:"page,omitempty" jsonschema:"页码，默认 1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"每页条数，默认 20，最大 100"`
}

// ListMediaOut list_media 出参。
type ListMediaOut struct {
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Items []res.MediaRes `json:"items"`
}

// listMedia 分页查询媒体库文件列表。
func (s *Service) listMedia(ctx context.Context, _ *sdk.CallToolRequest, in ListMediaIn) (*sdk.CallToolResult, res.MediaListRes, error) {
	if in.PageSize <= 0 {
		in.PageSize = 20
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	page, err := s.mediaLogic.GetList(ctx, &req.MediaListReq{
		PageReq:  req.PageReq{Page: in.Page, PageSize: in.PageSize},
		FileType: in.FileType,
		Keyword:  stringPtr(in.Keyword),
		FolderID: stringPtr(in.FolderID),
	})
	if err != nil {
		return nil, res.MediaListRes{}, fmt.Errorf("查询媒体列表失败: %w", err)
	}
	return nil, *page, nil
}

// MediaFolderItem list_media_folders 出参条目（树拍平后的单个文件夹）。
// 递归的树形结构无法被 jsonschema 生成器表达，故输出扁平列表，
// 以 depth 与 path（完整祖先路径）还原层级，便于直接选取 folder_id。
type MediaFolderItem struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	ParentID   *string `json:"parent_id"`
	ModuleKey  *string `json:"module_key"`
	MediaCount int64   `json:"media_count"`
	Depth      int     `json:"depth"`
	Path       string  `json:"path"`
}

// ListMediaFoldersOut list_media_folders 出参。
type ListMediaFoldersOut struct {
	Items []MediaFolderItem `json:"items"`
}

// flattenFolderTree 深度优先拍平文件夹树，Path 为「祖先名 / 本名」的完整路径。
func flattenFolderTree(nodes []res.MediaFolderRes) []MediaFolderItem {
	items := make([]MediaFolderItem, 0)
	var walk func(nodes []res.MediaFolderRes, depth int, prefix string)
	walk = func(nodes []res.MediaFolderRes, depth int, prefix string) {
		for _, n := range nodes {
			path := n.Name
			if prefix != "" {
				path = prefix + " / " + n.Name
			}
			items = append(items, MediaFolderItem{
				ID:         n.ID,
				Name:       n.Name,
				ParentID:   n.ParentID,
				ModuleKey:  n.ModuleKey,
				MediaCount: n.MediaCount,
				Depth:      depth,
				Path:       path,
			})
			walk(n.Children, depth+1, path)
		}
	}
	walk(nodes, 0, "")
	return items
}

// listMediaFolders 查询媒体库文件夹（按内容模块组织的层级目录，已拍平并附完整路径与文件数）。
func (s *Service) listMediaFolders(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, ListMediaFoldersOut, error) {
	tree, err := s.mediaFolderLogic.Tree(ctx)
	if err != nil {
		return nil, ListMediaFoldersOut{}, fmt.Errorf("查询媒体文件夹失败: %w", err)
	}
	return nil, ListMediaFoldersOut{Items: flattenFolderTree(tree.List)}, nil
}

// MoveMediaIn move_media 入参。
type MoveMediaIn struct {
	MediaIDs []string `json:"media_ids" jsonschema:"要移动的媒体文件 ID 列表（1-200 个）"`
	FolderID string   `json:"folder_id,omitempty" jsonschema:"目标文件夹 ID；不传表示移动到根目录"`
}

// MoveMediaOut move_media 出参。
type MoveMediaOut struct {
	Moved bool `json:"moved"`
	Count int  `json:"count"`
}

// moveMedia 批量移动媒体文件到指定文件夹。
func (s *Service) moveMedia(ctx context.Context, _ *sdk.CallToolRequest, in MoveMediaIn) (*sdk.CallToolResult, MoveMediaOut, error) {
	if len(in.MediaIDs) == 0 {
		return nil, MoveMediaOut{}, fmt.Errorf("media_ids 不能为空")
	}
	if len(in.MediaIDs) > 200 {
		return nil, MoveMediaOut{}, fmt.Errorf("media_ids 单次最多支持 200 个")
	}
	if err := s.mediaLogic.MoveMedia(ctx, in.MediaIDs, stringPtr(in.FolderID)); err != nil {
		return nil, MoveMediaOut{}, fmt.Errorf("移动媒体文件失败: %w", err)
	}
	return nil, MoveMediaOut{Moved: true, Count: len(in.MediaIDs)}, nil
}

// DeleteMediaIn delete_media 入参。
type DeleteMediaIn struct {
	MediaID string `json:"media_id" jsonschema:"要删除的媒体文件 ID"`
	Force   bool   `json:"force,omitempty" jsonschema:"文件被内容引用时需传 true 强制删除"`
	Confirm bool   `json:"confirm" jsonschema:"必须显式传 true 才会执行删除"`
}

// DeleteMediaOut delete_media 出参。
type DeleteMediaOut struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// deleteMedia 删除媒体文件（硬删除，含预设产物）。
func (s *Service) deleteMedia(ctx context.Context, _ *sdk.CallToolRequest, in DeleteMediaIn) (*sdk.CallToolResult, DeleteMediaOut, error) {
	if in.MediaID == "" {
		return nil, DeleteMediaOut{}, fmt.Errorf("media_id 不能为空")
	}
	if !in.Confirm {
		return nil, DeleteMediaOut{}, fmt.Errorf("删除是不可逆操作（将同时删除其全部预设产物），请确认后显式传 confirm=true 再执行；若文件正被内容引用需先传 force=true")
	}
	if err := s.mediaLogic.Delete(ctx, in.MediaID, in.Force); err != nil {
		return nil, DeleteMediaOut{}, fmt.Errorf("删除媒体文件失败: %w", err)
	}
	return nil, DeleteMediaOut{Deleted: true, ID: in.MediaID}, nil
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
