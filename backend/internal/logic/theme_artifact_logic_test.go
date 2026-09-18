package logic

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"novablog/enum"
)

// isBizCode 判断错误是否为指定业务码。
func isBizCode(err error, code int) bool {
	var bizErr *enum.BizError
	return errors.As(err, &bizErr) && bizErr.Code == code
}

// buildTestArchive 构造测试用 tar.gz（内存内打包）。
func buildTestArchive(t *testing.T, entries []tar.Header, contents map[string][]byte) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, h := range entries {
		if h.Typeflag == tar.TypeReg {
			h.Size = int64(len(contents[h.Name]))
		}
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatalf("写 tar 头失败: %v", err)
		}
		if h.Typeflag == tar.TypeReg {
			if _, err := tw.Write(contents[h.Name]); err != nil {
				t.Fatalf("写 tar 内容失败: %v", err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("关闭 tar 失败: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("关闭 gzip 失败: %v", err)
	}

	path := filepath.Join(t.TempDir(), "artifact.tar.gz")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("写测试归档失败: %v", err)
	}
	return path
}

// TestSafeExtractNormal 正常制品解压成功。
func TestSafeExtractNormal(t *testing.T) {
	archive := buildTestArchive(t,
		[]tar.Header{
			{Name: "theme.json", Typeflag: tar.TypeReg},
			{Name: "dist", Typeflag: tar.TypeDir},
			{Name: "dist/index.html", Typeflag: tar.TypeReg},
			{Name: "dist/_next/static/app.js", Typeflag: tar.TypeReg},
		},
		map[string][]byte{
			"theme.json":               []byte(`{"id":"demo"}`),
			"dist/index.html":          []byte("<html></html>"),
			"dist/_next/static/app.js": []byte("console.log(1)"),
		},
	)

	dest := t.TempDir()
	logic := NewThemeArtifactLogic()
	if err := logic.SafeExtract(archive, dest); err != nil {
		t.Fatalf("正常解压应成功: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dest, "dist", "_next", "static", "app.js"))
	if err != nil || string(data) != "console.log(1)" {
		t.Fatalf("解压内容不正确: %v %q", err, string(data))
	}
}

// TestSafeExtractTarSlip 路径穿越条目必须被拒绝。
func TestSafeExtractTarSlip(t *testing.T) {
	archive := buildTestArchive(t,
		[]tar.Header{
			{Name: "../evil.txt", Typeflag: tar.TypeReg},
		},
		map[string][]byte{"../evil.txt": []byte("evil")},
	)

	dest := t.TempDir()
	if err := NewThemeArtifactLogic().SafeExtract(archive, dest); err == nil {
		t.Fatal("包含 .. 的条目应被拒绝")
	}
	if _, err := os.Stat(filepath.Join(dest, "..", "evil.txt")); err == nil {
		t.Fatal("穿越文件不应被写出")
	}
}

// TestSafeExtractSymlink 符号链接条目必须被拒绝。
func TestSafeExtractSymlink(t *testing.T) {
	archive := buildTestArchive(t,
		[]tar.Header{
			{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"},
		},
		nil,
	)

	if err := NewThemeArtifactLogic().SafeExtract(archive, t.TempDir()); err == nil {
		t.Fatal("符号链接条目应被拒绝")
	}
}

// writeManifestDir 构造一个制品目录（theme.json + dist）。
func writeManifestDir(t *testing.T, manifest string, withDist bool) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "theme.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if withDist {
		if err := os.MkdirAll(filepath.Join(dir, "dist"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "dist", "index.html"), []byte("<html></html>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestLoadManifest 清单校验：合法/引擎不支持/API 版本不兼容/缺 dist。
func TestLoadManifest(t *testing.T) {
	logic := NewThemeArtifactLogic()
	valid := `{"id":"tech-geek","name":"极客风","version":"1.0.0","engine":"next-static","api_compat":"v1"}`

	if _, err := logic.LoadManifest(writeManifestDir(t, valid, true)); err != nil {
		t.Fatalf("合法清单应通过校验: %v", err)
	}

	if _, err := logic.LoadManifest(writeManifestDir(t, valid, false)); !isBizCode(err, enum.ErrThemeManifestInvalid.Code) {
		t.Fatalf("缺少 dist 应报清单校验失败: %v", err)
	}

	noEngine := `{"id":"demo","name":"d","version":"1.0.0","api_compat":"v1"}`
	if _, err := logic.LoadManifest(writeManifestDir(t, noEngine, true)); !errors.Is(err, enum.ErrThemeEngineUnsupported) {
		t.Fatalf("缺少引擎应报引擎不支持: %v", err)
	}

	ssrEngine := `{"id":"demo","name":"d","version":"1.0.0","engine":"next-standalone","api_compat":"v1"}`
	if _, err := logic.LoadManifest(writeManifestDir(t, ssrEngine, true)); !errors.Is(err, enum.ErrThemeEngineUnsupported) {
		t.Fatalf("SSR 引擎应被拒绝: %v", err)
	}

	badCompat := `{"id":"demo","name":"d","version":"1.0.0","engine":"next-static","api_compat":"v2"}`
	if _, err := logic.LoadManifest(writeManifestDir(t, badCompat, true)); !errors.Is(err, enum.ErrThemeAPICompatIncompatible) {
		t.Fatalf("不兼容 API 版本应被拒绝: %v", err)
	}
}

// TestPlaceArtifact 原子落盘：成功、冲突。
func TestPlaceArtifact(t *testing.T) {
	dataDir := t.TempDir()
	old := getThemeSettings()
	defer SetThemeSettings(old)
	SetThemeSettings(&ThemeSettings{DataDir: dataDir, MaxArtifactMB: 100})

	logic := NewThemeArtifactLogic()
	tmpRoot, err := logic.TempRoot()
	if err != nil {
		t.Fatalf("创建临时根失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpRoot, "index.html"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	rel, err := logic.PlaceArtifact(tmpRoot, "demo", "1.0.0")
	if err != nil {
		t.Fatalf("落盘应成功: %v", err)
	}
	if rel != "demo-1.0.0" {
		t.Fatalf("相对目录名不正确: %q", rel)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "demo-1.0.0", "index.html")); err != nil {
		t.Fatalf("落盘文件不存在: %v", err)
	}

	// 临时根已 mv 走，新建一个再试 → 冲突
	tmpRoot2, _ := logic.TempRoot()
	defer os.RemoveAll(tmpRoot2)
	if _, err := logic.PlaceArtifact(tmpRoot2, "demo", "1.0.0"); !errors.Is(err, enum.ErrThemeVersionExists) {
		t.Fatalf("同版本重复落盘应返回冲突: %v", err)
	}
}

// TestMatchFallback fallback 最长前缀匹配。
func TestMatchFallback(t *testing.T) {
	manifest := &ThemeManifest{Routes: manifestRoutes{Fallback: map[string]string{
		"/articles/*": "/article.html",
		"/travel/*":   "/travel.html",
	}}}

	if shell, ok := matchFallback(manifest, "/articles/hello-world"); !ok || shell != "/article.html" {
		t.Fatalf("文章路径应回退到 article.html: %q %v", shell, ok)
	}
	if _, ok := matchFallback(manifest, "/music/list"); ok {
		t.Fatal("未匹配路径不应有 fallback")
	}
	if _, ok := matchFallback(&ThemeManifest{}, "/articles/a"); ok {
		t.Fatal("空清单不应有 fallback")
	}
}

// TestContainsDotSegment dotfile 拒绝。
func TestContainsDotSegment(t *testing.T) {
	if !containsDotSegment(".env") || !containsDotSegment("dist/.hidden/x") {
		t.Fatal("dotfile 应被识别")
	}
	if containsDotSegment("dist/index.html") {
		t.Fatal("正常路径不应被识别为 dotfile")
	}
}
