package updater

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zeroReader 持续产出零字节, 用于流式构造超大响应
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// 下载成功: 内容完整落盘
func TestDownloadFileWritesContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "archive-bytes-123")
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "asset.tar.gz")
	if err := downloadFile(ts.URL, dest); err != nil {
		t.Fatalf("downloadFile failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "archive-bytes-123" {
		t.Errorf("downloaded content = %q, want %q", got, "archive-bytes-123")
	}
}

// 产物超过 100MB 上限时中止下载
func TestDownloadFileEnforcesSizeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("流式 100MB 较慢, short 模式跳过")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.CopyN(w, zeroReader{}, MaxDownloadBodySize+1)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "asset.tar.gz")
	err := downloadFile(ts.URL, dest)
	if err == nil {
		t.Fatal("downloadFile should abort oversized download, got nil")
	}
	if !strings.Contains(err.Error(), "100MB") {
		t.Errorf("error should mention the size limit, got: %v", err)
	}
}

// 恰好等于上限的产物可完整下载
func TestDownloadFileAcceptsExactLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("流式 100MB 较慢, short 模式跳过")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.CopyN(w, zeroReader{}, MaxDownloadBodySize)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "asset.tar.gz")
	if err := downloadFile(ts.URL, dest); err != nil {
		t.Fatalf("downloadFile should accept download at exact limit, got: %v", err)
	}
}

// HTTP 错误状态返回清晰错误
func TestDownloadFileHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer ts.Close()

	err := downloadFile(ts.URL+"/missing", filepath.Join(t.TempDir(), "asset"))
	if err == nil {
		t.Fatal("downloadFile should fail on HTTP error, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should carry the HTTP status, got: %v", err)
	}
}

// Release 资产下载会经历 302 跳转 (GitHub 指向对象存储), 跳转后仍能取到内容
func TestDownloadFileFollowsRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/asset.tar.gz" {
			_, _ = fmt.Fprint(w, "redirected-content")
			return
		}
		http.Redirect(w, r, "/asset.tar.gz", http.StatusFound)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "asset.tar.gz")
	if err := downloadFile(ts.URL+"/redirect", dest); err != nil {
		t.Fatalf("downloadFile failed: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "redirected-content" {
		t.Errorf("downloaded content = %q, want %q", got, "redirected-content")
	}
}
