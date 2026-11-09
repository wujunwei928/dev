package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newAssetServer 起伪装 Releases API 的测试服务。
// paths 映射请求路径 → (HTTP 状态码, release JSON);记录收到的请求路径供断言
func newAssetServer(t *testing.T, paths map[string]int, releases map[string]releaseInfo) (*httptest.Server, *[]string) {
	t.Helper()
	var got []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.URL.Path)
		code, ok := paths[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if rel, ok := releases[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(rel)
			return
		}
		w.WriteHeader(code)
	}))
	t.Cleanup(ts.Close)
	return ts, &got
}

func mustJSON(t *testing.T, rel releaseInfo) string {
	t.Helper()
	b, err := json.Marshal(rel)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// latest 端点返回完整 release (tag + 资产列表)
func TestFetchReleaseLatest(t *testing.T) {
	rel := releaseInfo{
		TagName: "v0.2.1",
		Assets: []releaseAsset{
			{Name: "checksums.txt", BrowserDownloadURL: "https://example/c.sum"},
		},
	}
	ts, got := newAssetServer(t,
		map[string]int{"/repos/wujunwei928/dev/releases/latest": 200},
		map[string]releaseInfo{"/repos/wujunwei928/dev/releases/latest": rel},
	)

	res, err := fetchRelease(ts.URL, "")
	if err != nil {
		t.Fatalf("fetchRelease failed: %v", err)
	}
	if res.TagName != "v0.2.1" {
		t.Errorf("TagName = %q, want v0.2.1", res.TagName)
	}
	if len(res.Assets) != 1 || res.Assets[0].Name != "checksums.txt" {
		t.Errorf("Assets = %+v, want checksums.txt", res.Assets)
	}
	if len(*got) != 1 || (*got)[0] != "/repos/wujunwei928/dev/releases/latest" {
		t.Errorf("unexpected request paths: %v", *got)
	}
}

// 指定 tag 时查询 tags 端点
func TestFetchReleaseByTag(t *testing.T) {
	rel := releaseInfo{TagName: "v0.1.5"}
	path := "/repos/wujunwei928/dev/releases/tags/v0.1.5"
	ts, _ := newAssetServer(t, map[string]int{path: 200}, map[string]releaseInfo{path: rel})

	res, err := fetchRelease(ts.URL, "v0.1.5")
	if err != nil {
		t.Fatalf("fetchRelease failed: %v", err)
	}
	if res.TagName != "v0.1.5" {
		t.Errorf("TagName = %q, want v0.1.5", res.TagName)
	}
}

// 用户带 v 前缀查询而实际 tag 无前缀 (历史 Release 形态) 时自动回落
func TestFetchReleaseByTagFallsBackWithoutV(t *testing.T) {
	pathNoV := "/repos/wujunwei928/dev/releases/tags/0.1.3"
	ts, got := newAssetServer(t,
		map[string]int{
			"/repos/wujunwei928/dev/releases/tags/v0.1.3": 404,
			pathNoV: 200,
		},
		map[string]releaseInfo{pathNoV: {TagName: "0.1.3"}},
	)

	res, err := fetchRelease(ts.URL, "v0.1.3")
	if err != nil {
		t.Fatalf("fetchRelease failed: %v", err)
	}
	if res.TagName != "0.1.3" {
		t.Errorf("TagName = %q, want 0.1.3", res.TagName)
	}
	if len(*got) != 2 {
		t.Errorf("expected exact-then-fallback requests, got: %v", *got)
	}
}

// 两种形态都查不到时明确报错
func TestFetchReleaseTagNotFound(t *testing.T) {
	ts, _ := newAssetServer(t, nil, nil)

	_, err := fetchRelease(ts.URL, "v9.9.9")
	if err == nil {
		t.Fatal("fetchRelease should fail on unknown tag, got nil")
	}
	if !strings.Contains(err.Error(), "v9.9.9") {
		t.Errorf("error should name the requested tag, got: %v", err)
	}
}

// 服务端错误同样返回清晰错误
func TestFetchReleaseServerError(t *testing.T) {
	path := "/repos/wujunwei928/dev/releases/latest"
	ts, _ := newAssetServer(t, map[string]int{path: 500}, nil)

	_, err := fetchRelease(ts.URL, "")
	if err == nil {
		t.Fatal("fetchRelease should fail on server error, got nil")
	}
}

// goreleaser 命名 (下划线, windows 用 zip) 按 GOOS/GOARCH 选资产
func TestPickAssetGoreleaserNaming(t *testing.T) {
	rel := &releaseInfo{TagName: "v0.2.1", Assets: []releaseAsset{
		{Name: "dev_0.2.1_linux_amd64.tar.gz", BrowserDownloadURL: "https://example/linux"},
		{Name: "dev_0.2.1_windows_amd64.zip", BrowserDownloadURL: "https://example/win"},
		{Name: "dev_0.2.1_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example/mac"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example/c.sum"},
	}}

	cases := []struct {
		goos, arch, wantURL string
	}{
		{"linux", "amd64", "https://example/linux"},
		{"windows", "amd64", "https://example/win"},
		{"darwin", "arm64", "https://example/mac"},
	}
	for _, c := range cases {
		a, err := pickAsset(rel, c.goos, c.arch)
		if err != nil {
			t.Fatalf("pickAsset(%s/%s) failed: %v", c.goos, c.arch, err)
		}
		if a.BrowserDownloadURL != c.wantURL {
			t.Errorf("pickAsset(%s/%s) = %s, want %s", c.goos, c.arch, a.BrowserDownloadURL, c.wantURL)
		}
	}
}

// tag 带 v 而产物名无 v (goreleaser 用去 v 版本号) 时仍能匹配
func TestPickAssetStripsVFromTagName(t *testing.T) {
	rel := &releaseInfo{TagName: "v0.2.1", Assets: []releaseAsset{
		{Name: "dev_0.2.1_windows_amd64.zip", BrowserDownloadURL: "https://example/win"},
		{Name: "dev_0.2.1_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example/mac"},
	}}

	a, err := pickAsset(rel, "windows", "amd64")
	if err != nil {
		t.Fatalf("pickAsset failed: %v", err)
	}
	if a.BrowserDownloadURL != "https://example/win" {
		t.Errorf("asset = %s, want windows zip", a.BrowserDownloadURL)
	}

	a, err = pickAsset(rel, "darwin", "arm64")
	if err != nil {
		t.Fatalf("pickAsset failed: %v", err)
	}
	if a.BrowserDownloadURL != "https://example/mac" {
		t.Errorf("asset = %s, want darwin tar.gz", a.BrowserDownloadURL)
	}
}

// 不含当前平台产物时报错并说明平台
func TestPickAssetNoMatch(t *testing.T) {
	rel := &releaseInfo{TagName: "v0.2.1", Assets: []releaseAsset{
		{Name: "dev_0.2.1_linux_amd64.tar.gz", BrowserDownloadURL: "https://example/linux"},
	}}

	goos, arch := "plan9", "arm64"
	_, err := pickAsset(rel, goos, arch)
	if err == nil {
		t.Fatal("pickAsset should fail when no asset matches, got nil")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%s/%s", goos, arch)) {
		t.Errorf("error should mention platform %s/%s, got: %v", goos, arch, err)
	}
}

// checksums.txt 产物查找: 存在与缺失两种形态
func TestFindChecksumsAsset(t *testing.T) {
	rel := &releaseInfo{Assets: []releaseAsset{
		{Name: "dev_0.2.1_windows_amd64.zip"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example/c.sum"},
	}}
	a, ok := findChecksumsAsset(rel)
	if !ok || a.BrowserDownloadURL != "https://example/c.sum" {
		t.Errorf("findChecksumsAsset = (%+v, %v), want checksums.txt asset", a, ok)
	}

	empty := &releaseInfo{Assets: []releaseAsset{{Name: "dev-0.1.3-linux-amd64.tar.gz"}}}
	if _, ok := findChecksumsAsset(empty); ok {
		t.Error("findChecksumsAsset should report missing when release has no checksums")
	}
}
