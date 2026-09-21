package logic

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"novablog/enum"
	"novablog/pkg/proxyhttp"

	"go.uber.org/zap"
)

// ThemeManifest 主题包清单（theme.json），规范见《主题模板皮肤系统设计文档》3.2。
type ThemeManifest struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Engine      string         `json:"engine"`
	APICompat   string         `json:"api_compat"`
	Author      string         `json:"author"`
	Description string         `json:"description"`
	Screenshots []string       `json:"screenshots"`
	Routes      manifestRoutes `json:"routes"`
}

// manifestRoutes 路由声明（动态内容壳页面映射）。
type manifestRoutes struct {
	Fallback map[string]string `json:"fallback"`
}

// supportedEngines 支持的主题引擎白名单（v1 仅静态导出）。
var supportedEngines = map[string]bool{"next-static": true}

// supportedAPICompat 当前系统兼容的公开 API 版本。
const supportedAPICompat = "v1"

// themeIDPattern 主题标识格式：小写字母/数字/连字符。
var themeIDPattern = regexp.MustCompile(`^[a-z0-9-]{2,50}$`)

// versionPattern 语义化版本：x.y.z[-pre]。
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)

// 解压安全配额。
const (
	maxExtractTotalBytes = 200 << 20 // 总解压大小 ≤ 200MB
	maxExtractFiles      = 5000      // 文件数 ≤ 5000
)

// artifactHTTPTimeout 制品下载超时（长超时，支持大文件流式下载）。
const artifactHTTPTimeout = 10 * time.Minute

// ThemeArtifactLogic 主题制品处理：下载/校验/解压/注入/落盘。
type ThemeArtifactLogic struct{}

// NewThemeArtifactLogic 创建 ThemeArtifactLogic 实例。
func NewThemeArtifactLogic() *ThemeArtifactLogic {
	return &ThemeArtifactLogic{}
}

// downloadClient 按当前配置构建制品下载客户端：
// 配置了代理时走代理（探测缓存 + 代理中途失败自动降级直连），未配置则纯直连。
func (l *ThemeArtifactLogic) downloadClient() *http.Client {
	settings := getThemeSettings()
	return proxyhttp.NewResilient(proxyhttp.Config{
		Timeout:  artifactHTTPTimeout,
		ProxyURL: settings.ProxyURL,
		OnFallback: func(err error) {
			themeLog().Warn("主题下载代理不可用，已自动降级直连",
				zap.String("proxy", settings.ProxyURL), zap.Error(err))
		},
	})
}

// DownloadArtifact 流式下载制品到临时文件，边下边算 sha256。
// 返回临时文件路径与 sha256 hex（调用方负责清理临时文件）。
func (l *ThemeArtifactLogic) DownloadArtifact(ctx context.Context, rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", "", enum.ErrThemeManifestInvalid
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("构建制品下载请求失败: %w", err)
	}

	resp, err := l.downloadClient().Do(req)
	if err != nil {
		return "", "", fmt.Errorf("下载主题制品失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("下载主题制品失败(HTTP %d)", resp.StatusCode)
	}

	maxBytes := int64(getThemeSettings().MaxArtifactMB) << 20
	if maxBytes <= 0 {
		maxBytes = 100 << 20
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "nova-theme-*.tar.gz")
	if err != nil {
		return "", "", fmt.Errorf("创建临时文件失败: %w", err)
	}

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hash), io.LimitReader(resp.Body, maxBytes+1))
	closeErr := tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", "", fmt.Errorf("下载主题制品失败: %w", err)
	}
	if closeErr != nil {
		os.Remove(tmpFile.Name())
		return "", "", fmt.Errorf("写入临时文件失败: %w", closeErr)
	}
	if written > maxBytes {
		os.Remove(tmpFile.Name())
		return "", "", fmt.Errorf("制品超过大小上限 %dMB", getThemeSettings().MaxArtifactMB)
	}

	return tmpFile.Name(), hex.EncodeToString(hash.Sum(nil)), nil
}

// SafeExtract 安全解压 tar.gz 到 destDir（tar-slip 防护 + 大小/文件数配额）。
func (l *ThemeArtifactLogic) SafeExtract(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("打开制品文件失败: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("制品不是合法的 tar.gz 文件: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	var totalBytes int64
	var fileCount int
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("读取制品条目失败: %w", err)
		}

		// 目录穿越防护：拒绝绝对路径与 .. 逃逸
		name := filepath.ToSlash(header.Name)
		if strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
			return enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "制品包含非法路径条目: "+name, enum.ErrThemeManifestInvalid.HttpCode)
		}
		target := filepath.Join(destDir, filepath.FromSlash(name))
		if rel, relErr := filepath.Rel(destDir, target); relErr != nil || strings.HasPrefix(rel, "..") {
			return enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "制品包含非法路径条目: "+name, enum.ErrThemeManifestInvalid.HttpCode)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("创建制品目录失败: %w", err)
			}
		case tar.TypeReg:
			fileCount++
			if fileCount > maxExtractFiles {
				return enum.ErrThemeManifestInvalid
			}
			totalBytes += header.Size
			if totalBytes > maxExtractTotalBytes {
				return enum.ErrThemeManifestInvalid
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("创建制品目录失败: %w", err)
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
			if err != nil {
				return fmt.Errorf("写入制品文件失败: %w", err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return fmt.Errorf("解压制品文件失败: %w", err)
			}
			out.Close()
		case tar.TypeSymlink, tar.TypeLink:
			return enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "制品包含符号链接条目，已拒绝安装", enum.ErrThemeManifestInvalid.HttpCode)
		default:
			// 其余类型（FIFO、设备等）一律跳过
		}
	}
	return nil
}

// LoadManifest 读取并校验制品根目录的 theme.json。
func (l *ThemeArtifactLogic) LoadManifest(rootDir string) (*ThemeManifest, error) {
	data, err := os.ReadFile(filepath.Join(rootDir, "theme.json"))
	if err != nil {
		return nil, enum.ErrThemeManifestInvalid
	}

	var manifest ThemeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, enum.ErrThemeManifestInvalid
	}

	switch {
	case !themeIDPattern.MatchString(manifest.ID):
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "主题清单 id 不合法（需匹配 ^[a-z0-9-]{2,50}$）", enum.ErrThemeManifestInvalid.HttpCode)
	case strings.TrimSpace(manifest.Name) == "":
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "主题清单缺少 name", enum.ErrThemeManifestInvalid.HttpCode)
	case !versionPattern.MatchString(manifest.Version):
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "主题清单 version 不是合法的语义化版本", enum.ErrThemeManifestInvalid.HttpCode)
	case !supportedEngines[manifest.Engine]:
		return nil, enum.ErrThemeEngineUnsupported
	case manifest.APICompat != supportedAPICompat:
		return nil, enum.ErrThemeAPICompatIncompatible
	}

	if info, err := os.Stat(filepath.Join(rootDir, "dist", "index.html")); err != nil || info.IsDir() {
		return nil, enum.NewBizError(enum.ErrThemeManifestInvalid.Code, "制品缺少 dist/index.html，无法静态托管", enum.ErrThemeManifestInvalid.HttpCode)
	}
	return &manifest, nil
}

// InjectThemeConfig 跨域部署时向制品 dist 目录写入 theme-config.js
// （Web 根为 dist/，主题页面经 <script src="/theme-config.js"> 引用；同域不写，回退相对路径）。
func (l *ThemeArtifactLogic) InjectThemeConfig(rootDir, apiBase string) error {
	if strings.TrimSpace(apiBase) == "" {
		return nil
	}
	content := "// 由 CMS 部署端注入的运行时配置（主题更新/重装时自动重写）\nwindow.__NOVA_CONFIG__ = { apiBase: " + fmt.Sprintf("%q", apiBase) + " };\n"
	return os.WriteFile(filepath.Join(rootDir, "dist", "theme-config.js"), []byte(content), 0o644)
}

// ApplyThemeConfig 将运行时 apiBase 应用到主题制品：非空写入 theme-config.js，空则清除已注入文件
// （后台修改公开 API 地址后对已安装主题重注入；清空表示回退同域相对路径取数）。
func (l *ThemeArtifactLogic) ApplyThemeConfig(rootDir, apiBase string) error {
	if strings.TrimSpace(apiBase) == "" {
		if err := os.Remove(filepath.Join(rootDir, "dist", "theme-config.js")); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return l.InjectThemeConfig(rootDir, apiBase)
}

// TempRoot 在 dataDir/.tmp 下创建临时解压根（与 dataDir 同一文件系统，保证可原子 mv）。
func (l *ThemeArtifactLogic) TempRoot() (string, error) {
	settings := getThemeSettings()
	tmpBase := filepath.Join(settings.DataDir, ".tmp")
	if err := os.MkdirAll(tmpBase, 0o755); err != nil {
		return "", fmt.Errorf("创建主题临时目录失败: %w", err)
	}
	return os.MkdirTemp(tmpBase, "install-")
}

// PlaceArtifact 将临时制品目录原子移动到 {dataDir}/{themeID}-{version}，返回相对目录名。
// 目标已存在时返回 ErrThemeVersionExists（由调用方决定覆盖）。
func (l *ThemeArtifactLogic) PlaceArtifact(tmpRoot, themeID, version string) (string, error) {
	settings := getThemeSettings()
	if err := os.MkdirAll(settings.DataDir, 0o755); err != nil {
		return "", fmt.Errorf("创建主题数据目录失败: %w", err)
	}

	target := filepath.Join(settings.DataDir, themeID+"-"+version)
	if _, err := os.Stat(target); err == nil {
		return "", enum.ErrThemeVersionExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("检查主题目录失败: %w", err)
	}

	if err := os.Rename(tmpRoot, target); err != nil {
		return "", fmt.Errorf("落盘主题制品失败: %w", err)
	}
	return themeID + "-" + version, nil
}

// 内置兜底主题相关功能已移除（由前端提示官方地址不可达，允许重试）
