package webdav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// multistatusBody 模拟坚果云/Alist 风格的 PROPFIND 响应（含 base path /dav）。
const multistatusBody = `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>/dav/</D:href>
    <D:propstat><D:prop><D:resourcetype><D:collection/></D:resourcetype><D:displayname>/</D:displayname></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>
  </D:response>
  <D:response>
    <D:href>/dav/%E7%85%A7%E7%89%87/</D:href>
    <D:propstat><D:prop><D:resourcetype><D:collection/></D:resourcetype><D:displayname>照片</D:displayname></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>
  </D:response>
  <D:response>
    <D:href>/dav/%E7%85%A7%E7%89%87/sea%20sunset.jpg</D:href>
    <D:propstat><D:prop><D:resourcetype/><D:getcontentlength>204800</D:getcontentlength><D:displayname>sea sunset.jpg</D:displayname></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>
  </D:response>
</D:multistatus>`

// TestList_ParsesMultistatus 验证 List 解析 multistatus：剥离 base path、排除目录自身、
// 目录在前文件在后不保证（按响应顺序），文件大小与扩展名正确。
func TestList_ParsesMultistatus(t *testing.T) {
	var gotDepth, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotDepth = r.Header.Get("Depth")
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(multistatusBody))
	}))
	defer srv.Close()

	client := New(srv.URL+"/dav", "user", "pass", "")
	files, next, err := client.List(context.Background(), "/", "", 50)
	require.NoError(t, err)
	assert.Empty(t, next)
	assert.Equal(t, "1", gotDepth)
	assert.True(t, strings.HasPrefix(gotAuth, "Basic "))

	require.Len(t, files, 2)

	dir := files[0]
	assert.Equal(t, "照片", dir.Name)
	assert.True(t, dir.IsDir)
	assert.Equal(t, "/照片", dir.FileID)
	assert.Equal(t, "/", dir.ParentID)

	file := files[1]
	assert.Equal(t, "sea sunset.jpg", file.Name)
	assert.False(t, file.IsDir)
	assert.Equal(t, "/照片/sea sunset.jpg", file.FileID)
	assert.Equal(t, int64(204800), file.Size)
	assert.Equal(t, ".jpg", file.Ext)
}

// TestList_ExcludesSelfNested 验证子目录列表排除目录自身。
func TestList_ExcludesSelfNested(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/照片") && !strings.HasSuffix(r.URL.Path, "/照片/") {
			t.Errorf("PROPFIND 应请求 /照片，实际 %s", r.URL.Path)
		}
		// 默认命名空间声明（无前缀形式，部分服务器如此）：namespace 为 DAV:，Go 解析器按 namespace 匹配
		_, _ = w.Write([]byte(`<?xml version="1.0"?><multistatus xmlns="DAV:">
<response><href>/dav/%E7%85%A7%E7%89%87/</href><propstat><prop><resourcetype><collection/></resourcetype></prop></propstat></response>
<response><href>/dav/%E7%85%A7%E7%89%87/a.png</href><propstat><prop><getcontentlength>10</getcontentlength></prop></propstat></response>
</multistatus>`))
	}))
	defer srv.Close()

	// multistatus 缺 D: 前缀（部分服务器如此）：namespace 仍为 DAV:，Go 解析器按 namespace 匹配
	client := New(srv.URL+"/dav", "u", "p", "")
	files, _, err := client.List(context.Background(), "/照片", "", 50)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "a.png", files[0].Name)
}

// TestDownload_BasicAuthAndSize 验证 Download 携带 Basic Auth 并透传流。
func TestDownload_BasicAuthAndSize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", username)
		assert.Equal(t, "pass", password)
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("jpeg-bytes-here"))
	}))
	defer srv.Close()

	client := New(srv.URL, "user", "pass", "")
	reader, size, contentType, err := client.Download(context.Background(), "/photo.jpg")
	require.NoError(t, err)
	defer reader.Close()

	buf := make([]byte, 64)
	n, _ := reader.Read(buf)
	assert.Equal(t, "jpeg-bytes-here", string(buf[:n]))
	assert.Equal(t, int64(15), size)
	assert.Equal(t, "image/jpeg", contentType)
}

// TestUserInfo_Unauthorized 验证鉴权失败返回可读错误。
func TestUserInfo_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := New(srv.URL, "u", "p", "")
	_, err := client.UserInfo(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "账号或密码错误")
}

// TestNormalizeDir 验证目录路径归一化。
func TestNormalizeDir(t *testing.T) {
	assert.Equal(t, "/", normalizeDir(""))
	assert.Equal(t, "/", normalizeDir("/"))
	assert.Equal(t, "/a/b", normalizeDir("a/b"))
	assert.Equal(t, "/a/b", normalizeDir("/a/b/"))
	assert.Equal(t, "/a/b", normalizeDir("/a/b//"))
}

// TestRelativePath 验证 href 剥离 base path。
func TestRelativePath(t *testing.T) {
	client := New("https://dav.example.com/dav", "u", "p", "")
	assert.Equal(t, "/", client.relativePath("/dav/"))
	assert.Equal(t, "/sub", client.relativePath("/dav/sub/"))
	assert.Equal(t, "/sub/a.jpg", client.relativePath("/dav/sub/a.jpg"))
	// href 为完整 URL 时取其路径
	assert.Equal(t, "/sub/a.jpg", client.relativePath("https://dav.example.com/dav/sub/a.jpg"))
}
