package novablogapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsOfficialDownloadURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		rawURL  string
		want    bool
	}{
		{"官方代理地址命中", "http://api.example.com", "http://api.example.com/api/v1/themes/42/download", true},
		{"base 已含 api 前缀", "http://api.example.com/api/v1", "http://api.example.com/api/v1/themes/42/download", true},
		{"容忍尾斜杠", "http://api.example.com", "http://api.example.com/api/v1/themes/42/download/", true},
		{"host 不同不命中", "http://api.example.com", "http://evil.example.com/api/v1/themes/42/download", false},
		{"GitHub tar.gz 直链不命中", "http://api.example.com", "https://github.com/a/b/releases/download/v1/a-1.0.0.tar.gz", false},
		{"GitHub 仓库目录不命中", "http://api.example.com", "https://github.com/a/b/tree/HEAD/dir", false},
		{"主题 id 非数字不命中", "http://api.example.com", "http://api.example.com/api/v1/themes/abc/download", false},
		{"非 download 端点不命中", "http://api.example.com", "http://api.example.com/api/v1/themes/42/like", false},
		{"非法地址不命中", "http://api.example.com", "::::", false},
		{"base 为空", "", "http://api.example.com/api/v1/themes/42/download", false},
		{"rawURL 为空", "http://api.example.com", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOfficialDownloadURL(tc.baseURL, tc.rawURL); got != tc.want {
				t.Fatalf("IsOfficialDownloadURL(%q, %q) = %v, want %v", tc.baseURL, tc.rawURL, got, tc.want)
			}
		})
	}
}

func TestResolveDownload(t *testing.T) {
	t.Run("302 解析 Location 并携带 Bearer Token", func(t *testing.T) {
		var gotAuth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.Header().Set("Location", "https://github.com/a/b/releases/download/v1/a-1.0.0.tar.gz")
			w.WriteHeader(http.StatusFound)
		}))
		defer srv.Close()

		location, err := ResolveDownload(context.Background(), srv.URL+"/api/v1/themes/42/download", "tok-1")
		if err != nil {
			t.Fatalf("ResolveDownload 返回错误: %v", err)
		}
		if gotAuth != "Bearer tok-1" {
			t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer tok-1")
		}
		want := "https://github.com/a/b/releases/download/v1/a-1.0.0.tar.gz"
		if location != want {
			t.Fatalf("location = %q, want %q", location, want)
		}
	})

	t.Run("相对 Location 绝对化", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", "/files/a.tar.gz")
			w.WriteHeader(http.StatusFound)
		}))
		defer srv.Close()

		location, err := ResolveDownload(context.Background(), srv.URL+"/api/v1/themes/1/download", "tok")
		if err != nil {
			t.Fatalf("ResolveDownload 返回错误: %v", err)
		}
		if location != srv.URL+"/files/a.tar.gz" {
			t.Fatalf("location = %q, want %q", location, srv.URL+"/files/a.tar.gz")
		}
	})

	t.Run("401 返回 AuthError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401,"message":"token has been revoked"}`))
		}))
		defer srv.Close()

		_, err := ResolveDownload(context.Background(), srv.URL+"/api/v1/themes/1/download", "bad")
		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Fatalf("期望 AuthError，得到 %v", err)
		}
		if authErr.Message != "token has been revoked" {
			t.Fatalf("message = %q", authErr.Message)
		}
	})

	t.Run("404 返回 APIError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":404,"message":"not found"}`))
		}))
		defer srv.Close()

		_, err := ResolveDownload(context.Background(), srv.URL+"/api/v1/themes/1/download", "tok")
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("期望 APIError，得到 %v", err)
		}
		if apiErr.Message != "not found" {
			t.Fatalf("message = %q", apiErr.Message)
		}
	})

	t.Run("未按预期重定向时报错", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("plain body"))
		}))
		defer srv.Close()

		if _, err := ResolveDownload(context.Background(), srv.URL+"/api/v1/themes/1/download", "tok"); err == nil {
			t.Fatal("期望返回错误，得到 nil")
		}
	})

	t.Run("空 token 直接拒绝", func(t *testing.T) {
		_, err := ResolveDownload(context.Background(), "http://api.example.com/api/v1/themes/1/download", "")
		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Fatalf("期望 AuthError，得到 %v", err)
		}
	})
}
