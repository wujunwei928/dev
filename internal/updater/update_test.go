package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// fakeSite 伪装完整更新源: Releases API + 资产下载端点, 全部零真实网络
type fakeSite struct {
	baseURL string
	mu      sync.Mutex
	dlHits  []string
}

// newFakeSite 启动伪站点。
// releases 以 API 子路径为键 ("latest" / "tags/v0.2.1"); assets 为产物名字节流,
// 资产下载地址自动注入为 <base>/dl/<name>
func newFakeSite(t *testing.T, releases map[string]releaseInfo, assets map[string][]byte) *fakeSite {
	t.Helper()
	site := &fakeSite{}
	mux := http.NewServeMux()

	mux.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dl/")
		site.mu.Lock()
		site.dlHits = append(site.dlHits, name)
		site.mu.Unlock()
		b, ok := assets[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(b)
	})

	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		sub := strings.TrimPrefix(r.URL.Path, "/repos/wujunwei928/dev/releases/")
		rel, ok := releases[sub]
		if !ok {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		for i := range rel.Assets {
			rel.Assets[i].BrowserDownloadURL = site.baseURL + "/dl/" + rel.Assets[i].Name
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	site.baseURL = ts.URL
	return site
}

// makeRelease 构造仅含产物名的 release 响应
func makeRelease(tag string, names ...string) releaseInfo {
	rel := releaseInfo{TagName: tag}
	for _, n := range names {
		rel.Assets = append(rel.Assets, releaseAsset{Name: n})
	}
	return rel
}

// selfBytes 当前测试二进制内容 (作为「官方发布的新二进制」打进伪归档)
func selfBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(selfPath(t))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// platformArchiveName 按当前平台惯例的归档产物名 (goreleaser 命名)
func platformArchiveName(tagNoV string) string {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("dev_%s_%s_%s%s", tagNoV, runtime.GOOS, runtime.GOARCH, ext)
}

// buildSelfArchive 将测试二进制打包为指定格式归档字节流
func buildSelfArchive(t *testing.T, name string) []byte {
	t.Helper()
	payload := []struct {
		Name  string
		Bytes []byte
	}{
		{Name: "README", Bytes: []byte("packaged docs")},
		{Name: binaryName(), Bytes: selfBytes(t)},
	}
	if strings.HasSuffix(name, ".zip") {
		return buildZip(t, payload)
	}
	return buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "README", Bytes: []byte("packaged docs")},
		{Name: binaryName(), Bytes: selfBytes(t)},
	})
}

// checksumsFor 以真实哈希为归档生成 checksums.txt 内容 (夹具数据, 非被测代码产出)
func checksumsFor(t *testing.T, name string, data []byte) string {
	t.Helper()
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) + "  " + name + "\n"
}

// installDir 构造伪装的「已安装目录」, 放置当前版本二进制, 返回 (dir, execPath)
func installDir(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	name := binaryName()
	execPath := filepath.Join(dir, name)
	copySelf(t, execPath)
	return dir, execPath
}

// assertInstallDirClean 断言安装目录仅剩现役二进制, 无暂存残留、无 .old 遗留
func assertInstallDirClean(t *testing.T, dir string, execPath string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(execPath) {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("install dir should only contain the binary, got: %v", names)
	}
}

// 端到端: release 用户从查询到 rename 交换一次完成, 临时产物全部清理
func TestUpdateEndToEnd(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{
			"latest": makeRelease(tag, archName, "checksums.txt"),
		},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "v0.2.1") // 新二进制冒烟应答目标版本

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.From != "0.2.0" || res.To != "v0.2.1" {
		t.Errorf("result = %q -> %q, want 0.2.0 -> v0.2.1", res.From, res.To)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}
	if !res.ChecksumVerified {
		t.Error("ChecksumVerified = false, want true (checksums.txt 已提供)")
	}

	// 交换完成: 原位是新二进制内容
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("exec binary not replaced (err=%v, equal=%v)", err, bytes.Equal(got, selfBytes(t)))
	}
	// 下载与解压的临时产物已清理
	assertInstallDirClean(t, dir, execPath)
	// 资产与校验文件确实下载过
	site.mu.Lock()
	defer site.mu.Unlock()
	if len(site.dlHits) != 2 {
		t.Errorf("downloaded assets = %v, want archive + checksums.txt", site.dlHits)
	}
}

// 历史 Release 形态端到端: 连字符命名, tag 无 v, 无 checksums.txt, windows 也是 tar.gz
func TestUpdateEndToEndLegacyRelease(t *testing.T) {
	tag := "0.1.3"
	archName := fmt.Sprintf("dev-%s-%s-%s.tar.gz", tag, runtime.GOOS, runtime.GOARCH)
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName)},
		map[string][]byte{archName: archBytes},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "0.1.3")

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.1.2"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.To != "0.1.3" {
		t.Errorf("To = %q, want 0.1.3", res.To)
	}
	if res.ChecksumVerified {
		t.Error("ChecksumVerified = true, want false (历史 Release 无 checksums.txt)")
	}
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("exec binary not replaced (err=%v)", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// 已是最新时不再下载, 原位不动
func TestUpdateAlreadyUpToDateSkipsDownload(t *testing.T) {
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease("v0.2.1", "dev_0.2.1_linux_amd64.tar.gz")},
		nil,
	)
	dir, execPath := installDir(t)

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "v0.2.1"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("UpToDate = false, want true (current v0.2.1 = latest v0.2.1)")
	}
	site.mu.Lock()
	defer site.mu.Unlock()
	if len(site.dlHits) != 0 {
		t.Errorf("assets downloaded = %v, want none", site.dlHits)
	}
	assertInstallDirClean(t, dir, execPath)
}

// 当前版本比 latest 还新时不回退 (semver 大于)
func TestUpdateNewerThanLatestSkipsDownload(t *testing.T) {
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease("v0.2.1", "dev_0.2.1_linux_amd64.tar.gz")},
		nil,
	)
	_, execPath := installDir(t)

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "v0.3.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("UpToDate = false, want true (current v0.3.0 > latest v0.2.1)")
	}
}

// 本地构建 (devel 形态无法比较) 时继续安装官方版
func TestUpdateLocalBuildProceeds(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "v0.2.1")

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodLocal, Version: "devel+0123456"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false (devel 版本应继续安装)")
	}
	if res.To != "v0.2.1" {
		t.Errorf("To = %q, want v0.2.1", res.To)
	}
	assertInstallDirClean(t, dir, execPath)
}

// --version 指定版本可降级安装 (当前比目标新也执行替换)
func TestUpdateDowngradeByTargetTag(t *testing.T) {
	tag := "v0.1.5"
	archName := platformArchiveName("0.1.5")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"tags/v0.1.5": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "v0.1.5")

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "v0.2.1"},
		Target:   "v0.1.5",
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.To != "v0.1.5" || res.UpToDate {
		t.Errorf("result = %+v, want downgrade to v0.1.5", res)
	}
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("exec binary not replaced (err=%v)", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// --version 指定的版本不存在时报错
func TestUpdateTargetTagNotFound(t *testing.T) {
	site := newFakeSite(t, nil, nil)
	_, execPath := installDir(t)

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "v0.2.1"},
		Target:   "v9.9.9",
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err == nil {
		t.Fatal("Update should fail on unknown target tag, got nil")
	}
	if !strings.Contains(err.Error(), "v9.9.9") {
		t.Errorf("error should name the requested tag, got: %v", err)
	}
}

// sha256 不匹配时拒绝安装, 旧二进制原位不动, 无半成品残留
func TestUpdateChecksumMismatchKeepsOldBinary(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	// 独立错误字面量作为期望哈希, 制造不匹配
	badChecksums := otherSHA256 + "  " + archName + "\n"
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(badChecksums),
		},
	)
	dir, execPath := installDir(t)

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err == nil {
		t.Fatal("Update should fail on checksum mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "sha256") {
		t.Errorf("error should mention sha256, got: %v", err)
	}
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("old binary must stay untouched (err=%v)", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// 冒烟测试失败时拒绝替换, 旧版原位不动
func TestUpdateSmokeFailureKeepsOldBinary(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "v0.9.9") // 新二进制应答了错误版本

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err == nil {
		t.Fatal("Update should fail on smoke test mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "冒烟") {
		t.Errorf("error should mention smoke test, got: %v", err)
	}
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("old binary must stay untouched (err=%v)", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// 上次更新遗留的 .old 在本次更新启动时清理
func TestUpdateCleansStaleOldBinary(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:      archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	stale := filepath.Join(dir, binaryName()+".old")
	if err := os.WriteFile(stale, []byte("stale-leftover"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(fakeVersionEnv, "v0.2.1")

	if _, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale .old should be cleaned, stat err: %v", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// Release 不含当前平台产物时报错
func TestUpdateMissingPlatformAsset(t *testing.T) {
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease("v0.2.1", "dev_0.2.1_plan9_arm64.tar.gz")},
		nil,
	)
	_, execPath := installDir(t)

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err == nil {
		t.Fatal("Update should fail when no platform asset exists, got nil")
	}
	if !strings.Contains(err.Error(), runtime.GOOS) {
		t.Errorf("error should mention current platform, got: %v", err)
	}
}

// API 不可用时返回清晰错误, 旧版不动
func TestUpdateAPIFailureKeepsOldBinary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	ts.Close() // 直接关闭制造连接失败
	dir, execPath := installDir(t)

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodRelease, Version: "0.2.0"},
		APIBase:  ts.URL,
		ExecPath: execPath,
	})
	if err == nil {
		t.Fatal("Update should fail when API unreachable, got nil")
	}
	if !strings.Contains(err.Error(), "查询版本") {
		t.Errorf("error should state the query failed, got: %v", err)
	}
	assertInstallDirClean(t, dir, execPath)
}

// go install 构建的委托路径与回退行为见 goinstall_test.go (#9): Update 按安装方式分发,
// 下载替换路径经分发进入, 不再拒绝 go install 构建
