package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// runRoot 通过真实 cobra 入口执行 dev 命令, 返回标准输出与错误。
// 配置指向临时文件, 避免触碰 $HOME 下的真实配置
func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "dev.yaml")
	if err := os.WriteFile(cfg, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs(append([]string{"--config", cfg}, args...))
	err := rootCmd.Execute()
	return out.String(), err
}

// setAPIBase 注入 GitHub API base 到测试服务地址, 并在测试结束后还原
func setAPIBase(t *testing.T, url string) {
	t.Helper()
	viper.Set(UpdateConfigAPIBase, url)
	t.Cleanup(func() {
		viper.Set(UpdateConfigAPIBase, "")
	})
}

// newReleaseServer 起一个伪装 GitHub Releases API 的测试服务, 返回最新 tag
func newReleaseServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": tag})
	}))
	t.Cleanup(ts.Close)
	return ts
}

// 裸 dev update (release 构建) 已是最新时提示且不下载替换
func TestUpdateBareReleaseAlreadyUpToDate(t *testing.T) {
	ts := newReleaseServer(t, "v0.2.1")
	setAPIBase(t, ts.URL)
	Version = "v0.2.1"
	t.Cleanup(func() { Version = "" })

	out, err := runRoot(t, "update")
	if err != nil {
		t.Fatalf("bare update should not fail when up to date, got: %v", err)
	}
	if !strings.Contains(out, "already up to date") {
		t.Errorf("output should say already up to date, got: %q", out)
	}
}

// 本地构建裸更新先给出警告 (随后 API 失败则报错, 证明警告先于更新动作输出)
func TestUpdateBareLocalBuildWarnsAndFailsOnAPIDown(t *testing.T) {
	ts := newReleaseServer(t, "v0.2.1")
	ts.Close() // 先关闭制造网络失败
	setAPIBase(t, ts.URL)
	// Version 保持空串, 测试二进制经 buildinfo 解析为本地构建

	out, err := runRoot(t, "update")
	if err == nil {
		t.Fatal("bare update should fail when API is unreachable, got nil")
	}
	if !strings.Contains(out, "本地构建") {
		t.Errorf("output should warn about local build, got: %q", out)
	}
}

// release 构建下 --check 报告 current 与 latest (经 Releases API), cmd.Version 模拟 ldflags 注入
func TestUpdateCheckReportsCurrentAndLatest(t *testing.T) {
	ts := newReleaseServer(t, "v0.2.1")
	setAPIBase(t, ts.URL)
	Version = "0.2.0"
	t.Cleanup(func() { Version = "" })

	out, err := runRoot(t, "update", "--check")
	if err != nil {
		t.Fatalf("update --check should not fail, got: %v", err)
	}
	if !strings.Contains(out, "0.2.0") {
		t.Errorf("output should report current version 0.2.0, got: %q", out)
	}
	if !strings.Contains(out, "v0.2.1") {
		t.Errorf("output should report latest version v0.2.1, got: %q", out)
	}
	if strings.Contains(out, "already up to date") {
		t.Errorf("should not claim up to date, got: %q", out)
	}
}

// 已是最新时明确提示 already up to date
func TestUpdateCheckAlreadyUpToDate(t *testing.T) {
	ts := newReleaseServer(t, "v0.2.1")
	setAPIBase(t, ts.URL)
	Version = "v0.2.1"
	t.Cleanup(func() { Version = "" })

	out, err := runRoot(t, "update", "--check")
	if err != nil {
		t.Fatalf("update --check should not fail, got: %v", err)
	}
	if !strings.Contains(out, "already up to date") {
		t.Errorf("output should say already up to date, got: %q", out)
	}
}

// 网络失败时报错清晰, 且不影响其他子命令
func TestUpdateCheckFailureReturnsClearError(t *testing.T) {
	ts := newReleaseServer(t, "v0.2.1")
	ts.Close() // 先关闭制造网络失败
	setAPIBase(t, ts.URL)
	Version = "0.2.0"
	t.Cleanup(func() { Version = "" })

	_, err := runRoot(t, "update", "--check")
	if err == nil {
		t.Fatal("update --check should fail when API is unreachable, got nil")
	}
	if !strings.Contains(err.Error(), "查询版本") {
		t.Errorf("error should state the query failed clearly, got: %v", err)
	}

	// 失败不污染全局状态, 其他子命令仍可正常执行
	out, err := runRoot(t, "--version")
	if err != nil {
		t.Fatalf("--version should still work after a failed check, got: %v", err)
	}
	if out == "" {
		t.Error("--version should print something")
	}
}
