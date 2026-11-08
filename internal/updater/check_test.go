package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withGoShim 在 PATH 前置假 go: 应答 stdout 内容, errOutput 非空时写 stderr 并以 exitCode 退出。
// Windows 用 go.bat, 其余平台用带执行权限的 go 脚本。
func withGoShim(t *testing.T, stdout, errOutput string, exitCode int) {
	t.Helper()
	dir := t.TempDir()
	var content string
	if runtime.GOOS == "windows" {
		content = "@echo off\r\n"
		if stdout != "" {
			content += "echo " + stdout + "\r\n"
		}
		if errOutput != "" {
			content += "echo " + errOutput + " 1>&2\r\n"
		}
		if exitCode != 0 {
			content += fmt.Sprintf("exit /b %d\r\n", exitCode)
		}
	} else {
		content = "#!/bin/sh\n"
		if stdout != "" {
			content += "echo " + fmt.Sprintf("%q", stdout) + "\n"
		}
		if errOutput != "" {
			content += "echo " + fmt.Sprintf("%q", errOutput) + " >&2\n"
		}
		if exitCode != 0 {
			content += fmt.Sprintf("exit %d\n", exitCode)
		}
	}
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".bat"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// release 构建经 GitHub Releases API 查询最新版本
func TestCheckViaGitHubAPIReportsLatestRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/wujunwei928/dev/releases/latest" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.2.1"})
	}))
	defer ts.Close()

	res, err := Check(VersionInfo{Method: MethodRelease, Version: "0.2.0"}, ts.URL)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if res.Latest != "v0.2.1" {
		t.Errorf("Latest = %q, want v0.2.1", res.Latest)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}
}

// 本地构建同样走 GitHub Releases API
func TestCheckViaGitHubAPIForLocalBuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.2.1"})
	}))
	defer ts.Close()

	res, err := Check(VersionInfo{Method: MethodLocal, Version: "devel+0123456"}, ts.URL)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if res.Latest != "v0.2.1" {
		t.Errorf("Latest = %q, want v0.2.1", res.Latest)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}
}

// 当前版本与最新版本一致时提示已是最新
func TestCheckReportsUpToDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.2.1"})
	}))
	defer ts.Close()

	// release 构建注入的版本号无 v 前缀, 比较时须归一化
	res, err := Check(VersionInfo{Method: MethodRelease, Version: "0.2.1"}, ts.URL)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("UpToDate = false, want true (current %q vs latest %q)", res.Current, res.Latest)
	}
}

// API 不可达时返回清晰错误
func TestCheckAPIFailureReturnsClearError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := Check(VersionInfo{Method: MethodRelease, Version: "0.2.0"}, ts.URL)
	if err == nil {
		t.Fatal("Check should return error on API failure, got nil")
	}
	if !strings.Contains(err.Error(), "查询最新版本失败") {
		t.Errorf("error should state the query failed clearly, got: %v", err)
	}
}

// 连接层故障 (服务不可达) 同样返回清晰错误
func TestCheckAPIUnreachableReturnsClearError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := ts.URL
	ts.Close() // 先关掉服务制造连接失败

	_, err := Check(VersionInfo{Method: MethodRelease, Version: "0.2.0"}, url)
	if err == nil {
		t.Fatal("Check should return error when server is unreachable, got nil")
	}
	if !strings.Contains(err.Error(), "查询最新版本失败") {
		t.Errorf("error should state the query failed clearly, got: %v", err)
	}
}

// 空的 API base 使用默认 api.github.com
func TestCheckEmptyAPIBaseFallsBackToDefault(t *testing.T) {
	if defaultAPIBase != "https://api.github.com" {
		t.Fatalf("defaultAPIBase = %q, want https://api.github.com", defaultAPIBase)
	}
}

// go install 构建经 go list -m 查询最新版本 (PATH shim 假 go, 不依赖真实工具链与网络)
func TestCheckViaGoInstallShimReportsLatest(t *testing.T) {
	withGoShim(t, "github.com/wujunwei928/dev v0.3.0", "", 0)

	res, err := Check(VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"}, "")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if res.Latest != "v0.3.0" {
		t.Errorf("Latest = %q, want v0.3.0", res.Latest)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}
}

// go install 构建已是最新时 UpToDate 为真
func TestCheckGoInstallUpToDate(t *testing.T) {
	withGoShim(t, "github.com/wujunwei928/dev v0.2.0", "", 0)

	res, err := Check(VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"}, "")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("UpToDate = false, want true (latest %q)", res.Latest)
	}
}

// go 命令失败 (如网络受限) 时错误清晰并携带 stderr 上下文
func TestCheckGoInstallFailureReturnsClearError(t *testing.T) {
	withGoShim(t, "", "dial tcp: lookup proxy.golang.org: no such host", 1)

	_, err := Check(VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"}, "")
	if err == nil {
		t.Fatal("Check should return error when go list fails, got nil")
	}
	if !strings.Contains(err.Error(), "查询最新版本失败") {
		t.Errorf("error should state the query failed clearly, got: %v", err)
	}
	if !strings.Contains(err.Error(), "no such host") {
		t.Errorf("error should carry go's stderr context, got: %v", err)
	}
}

// go 不在 PATH 时错误清晰
func TestCheckGoMissingReturnsClearError(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := Check(VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"}, "")
	if err == nil {
		t.Fatal("Check should return error when go is not in PATH, got nil")
	}
	if !strings.Contains(err.Error(), "查询最新版本失败") {
		t.Errorf("error should state the query failed clearly, got: %v", err)
	}
}
