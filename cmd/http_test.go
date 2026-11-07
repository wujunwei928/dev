package cmd

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newMultipartUpload 构造上传文件的multipart请求体
func newMultipartUpload(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload-file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func doUpload(handler http.HandlerFunc, t *testing.T, filename, content string) *httptest.ResponseRecorder {
	t.Helper()
	body, contentType := newMultipartUpload(t, filename, content)
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestUploadHandlerWritesFile(t *testing.T) {
	dir := t.TempDir()
	handler := newUploadHandler(dir, false)

	rec := doUpload(handler, t, "hello.txt", "hello")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil || string(data) != "hello" {
		t.Errorf("uploaded file content mismatch: %q, err=%v", data, err)
	}
}

// 局域网内不可静默覆盖已存在文件
func TestUploadHandlerRejectsOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	handler := newUploadHandler(dir, false)
	rec := doUpload(handler, t, "exists.txt", "evil")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on overwrite attempt, got %d: %s", rec.Code, rec.Body.String())
	}
	data, _ := os.ReadFile(target)
	if string(data) != "original" {
		t.Errorf("existing file was modified: %q", data)
	}
}

func TestUploadHandlerOverwriteWhenAllowed(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	handler := newUploadHandler(dir, true)
	rec := doUpload(handler, t, "exists.txt", "new")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with overwrite enabled, got %d: %s", rec.Code, rec.Body.String())
	}
	data, _ := os.ReadFile(target)
	if string(data) != "new" {
		t.Errorf("file should be overwritten, got %q", data)
	}
}

// 文件名路径穿越必须被 basename 中和
func TestUploadHandlerNeutralizesPathTraversal(t *testing.T) {
	dir := t.TempDir()
	handler := newUploadHandler(dir, false)

	rec := doUpload(handler, t, `..\..\evil.txt`, "x")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "evil.txt")); err != nil {
		t.Errorf("file should be saved as basename in save dir: %v", err)
	}
}

// 未启用上传时, /upload 端点必须不存在
func TestRoutesWithoutUploadFlag(t *testing.T) {
	mux := http.NewServeMux()
	registerHttpRoutes(mux, httpRoutesOptions{StaticDir: t.TempDir(), UseUpload: false})

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(""))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("/upload should not be registered when upload disabled, got %d", rec.Code)
	}
}

// 上传页面不引入外部JS, 结果以textContent渲染防HTML注入
func TestIndexPageHasNoJquery(t *testing.T) {
	mux := http.NewServeMux()
	registerHttpRoutes(mux, httpRoutesOptions{StaticDir: t.TempDir(), UseUpload: true})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	page := rec.Body.String()
	if strings.Contains(page, "jquery") {
		t.Error("index page should not reference jquery")
	}
	if !strings.Contains(page, "textContent") {
		t.Error("upload result should render via textContent to avoid HTML injection")
	}
	if !strings.Contains(page, "/static") {
		t.Error("index page should link to /static download entry")
	}
}
