package updater

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// goShimLogEnv 录制型 go shim 记录调用参数的文件路径环境变量
const goShimLogEnv = "DEV_UPDATE_TEST_GO_SHIM_LOG"

// withRecordingGoShim 在 PATH 前置可录制调用的假 go:
// 每次调用把完整参数追加到日志文件; 参数首词为 list 时应答 listAnswer;
// exitCode 为非 list (即 install) 调用的退出码, list 调用恒成功。
// 返回日志文件路径。全程零真实网络。
func withRecordingGoShim(t *testing.T, listAnswer string, exitCode int) string {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "go-shim.log")
	if err := os.WriteFile(logPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(goShimLogEnv, logPath)

	var content string
	name := "go"
	if runtime.GOOS == "windows" {
		content = "@echo off\r\n"
		content += ">>\"%" + goShimLogEnv + "%\" echo %*\r\n"
		if listAnswer != "" {
			content += "if /i \"%~1\"==\"list\" (\r\n"
			content += "  echo " + listAnswer + "\r\n"
			if exitCode != 0 {
				content += ") else (\r\n"
				content += fmt.Sprintf("  exit /b %d\r\n", exitCode)
			}
			content += ")\r\n"
		} else if exitCode != 0 {
			content += fmt.Sprintf("exit /b %d\r\n", exitCode)
		}
		name += ".bat"
	} else {
		content = "#!/bin/sh\n"
		content += "printf '%s\\n' \"$*\" >> \"$" + goShimLogEnv + "\"\n"
		if listAnswer != "" {
			content += "if [ \"$1\" = \"list\" ]; then\n"
			content += "  echo '" + listAnswer + "'\n"
			content += "else\n"
			if exitCode != 0 {
				content += fmt.Sprintf("  exit %d\n", exitCode)
			}
			content += "fi\n"
		} else if exitCode != 0 {
			content += fmt.Sprintf("exit %d\n", exitCode)
		}
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
	return logPath
}

// shimCalls 读取 go shim 的调用参数日志, 每行一次调用
func shimCalls(t *testing.T, logPath string) []string {
	t.Helper()
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			calls = append(calls, line)
		}
	}
	return calls
}

// deadDownloadServer 起一个立即关闭的伪下载源: 下载路径若被触碰即连接失败
func deadDownloadServer(t *testing.T) string {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "download path must not be reached", http.StatusInternalServerError)
	}))
	ts.Close()
	return ts.URL
}

// go install 构建裸更新: 委托 go install <module>@latest, 事后实测版本真的变了
func TestUpdateGoInstallDelegatesLatest(t *testing.T) {
	logPath := withRecordingGoShim(t, "github.com/wujunwei928/dev v0.3.0", 0)
	t.Setenv(fakeVersionEnv, "v0.3.0") // shim 未真装, 实测应答目标版本

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  deadDownloadServer(t), // 下载路径不应被触碰
		ExecPath: selfPath(t),           // 事后实测对象即测试二进制自身
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if res.From != "v0.2.0" || res.To != "v0.3.0" {
		t.Errorf("result = %q -> %q, want v0.2.0 -> v0.3.0", res.From, res.To)
	}
	if res.Method != MethodGoInstall {
		t.Errorf("Method = %v, want MethodGoInstall", res.Method)
	}
	if res.UpToDate {
		t.Error("UpToDate = true, want false")
	}

	// 委托参数: 先经 go list 查最新版, 再 go install <module>@latest
	calls := shimCalls(t, logPath)
	want := []string{
		"list -m github.com/wujunwei928/dev@latest",
		"install github.com/wujunwei928/dev@latest",
	}
	if len(calls) != len(want) {
		t.Fatalf("go invoked %d time(s) %v, want %v", len(calls), calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("go call #%d = %q, want %q", i+1, calls[i], want[i])
		}
	}
}

// --version 指定版本经 go install <module>@vX.Y.Z 安装, 兼做降级, 无需预查最新版
func TestUpdateGoInstallDelegatesPinnedTarget(t *testing.T) {
	cases := []struct {
		name   string
		target string // 用户传入的 --version 值
	}{
		{"带 v 前缀", "v0.2.0"},
		{"不带 v 前缀", "0.2.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			logPath := withRecordingGoShim(t, "", 0)
			t.Setenv(fakeVersionEnv, "v0.2.0")

			res, err := Update(UpdateOptions{
				Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.3.0"}, // 当前更新, 显式降级
				Target:   c.target,
				APIBase:  deadDownloadServer(t),
				ExecPath: selfPath(t),
			})
			if err != nil {
				t.Fatalf("Update failed: %v", err)
			}
			if res.To != "v0.2.0" || res.From != "v0.3.0" {
				t.Errorf("result = %q -> %q, want v0.3.0 -> v0.2.0", res.From, res.To)
			}

			// 仅一次调用: go install <module>@v0.2.0 (指定版本不预查最新版)
			calls := shimCalls(t, logPath)
			if len(calls) != 1 {
				t.Fatalf("go invoked %d time(s) %v, want exactly 1 install", len(calls), calls)
			}
			if wantRef := "install github.com/wujunwei928/dev@v0.2.0"; calls[0] != wantRef {
				t.Errorf("go call = %q, want %q", calls[0], wantRef)
			}
		})
	}
}

// 已是最新版时短路返回 UpToDate, 不重复委托 go install
func TestUpdateGoInstallAlreadyUpToDateSkipsInstall(t *testing.T) {
	logPath := withRecordingGoShim(t, "github.com/wujunwei928/dev v0.2.0", 0)

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  deadDownloadServer(t),
		ExecPath: selfPath(t),
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.UpToDate {
		t.Errorf("UpToDate = false, want true (current v0.2.0 = latest v0.2.0)")
	}
	if res.To != "v0.2.0" || res.From != "v0.2.0" {
		t.Errorf("result = %q -> %q, want v0.2.0 -> v0.2.0", res.From, res.To)
	}

	// 仅一次 go list 预查, 无 go install 重装
	calls := shimCalls(t, logPath)
	if len(calls) != 1 || calls[0] != "list -m github.com/wujunwei928/dev@latest" {
		t.Fatalf("go calls = %v, want only [list -m github.com/wujunwei928/dev@latest]", calls)
	}
}

// go install 退出码为 0 但版本未变 (如 shim 未真装) 时, 事后实测报错, 不谎报成功
func TestUpdateGoInstallVersionUnchangedFails(t *testing.T) {
	withRecordingGoShim(t, "github.com/wujunwei928/dev v0.3.0", 0)
	t.Setenv(fakeVersionEnv, "v0.2.0") // 实测应答旧版本: 安装未生效

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  deadDownloadServer(t),
		ExecPath: selfPath(t),
	})
	if err == nil {
		t.Fatal("Update should fail when measured version is unchanged, got nil")
	}
	if !strings.Contains(err.Error(), "未变") {
		t.Errorf("error should state version did not change, got: %v", err)
	}
}

// 指定版本安装后实测版本与请求不符时报错
func TestUpdateGoInstallPinnedMismatchFails(t *testing.T) {
	withRecordingGoShim(t, "", 0)
	t.Setenv(fakeVersionEnv, "v0.9.9") // 实测应答了错误版本

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		Target:   "v0.2.0",
		APIBase:  deadDownloadServer(t),
		ExecPath: selfPath(t),
	})
	if err == nil {
		t.Fatal("Update should fail when measured version differs from target, got nil")
	}
	if !strings.Contains(err.Error(), "期望 v0.2.0") {
		t.Errorf("error should name the expected version, got: %v", err)
	}
}

// go install 执行失败 (如网络受限) 时错误清晰并携带委托 ref
func TestUpdateGoInstallFailureReturnsClearError(t *testing.T) {
	withRecordingGoShim(t, "github.com/wujunwei928/dev v0.3.0", 1)

	_, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  deadDownloadServer(t),
		ExecPath: selfPath(t),
	})
	if err == nil {
		t.Fatal("Update should fail when go install fails, got nil")
	}
	if !strings.Contains(err.Error(), "github.com/wujunwei928/dev@latest") {
		t.Errorf("error should name the delegated ref, got: %v", err)
	}
}

// go 不在 PATH 时自动回退下载替换路径并成功升级, 结果标注回退事实
func TestUpdateGoInstallFallsBackToDownloadWhenGoMissing(t *testing.T) {
	tag := "v0.2.1"
	archName := platformArchiveName("0.2.1")
	archBytes := buildSelfArchive(t, archName)
	site := newFakeSite(t,
		map[string]releaseInfo{"latest": makeRelease(tag, archName, "checksums.txt")},
		map[string][]byte{
			archName:        archBytes,
			"checksums.txt": []byte(checksumsFor(t, archName, archBytes)),
		},
	)
	dir, execPath := installDir(t)
	t.Setenv(fakeVersionEnv, "v0.2.1") // 新二进制冒烟应答目标版本
	t.Setenv("PATH", "")               // go 不可用

	res, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  site.baseURL,
		ExecPath: execPath,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.GoMissingFallback {
		t.Error("GoMissingFallback = false, want true")
	}
	if res.Method != MethodGoInstall {
		t.Errorf("Method = %v, want MethodGoInstall (按原安装方式标注)", res.Method)
	}
	if res.To != tag || res.UpToDate {
		t.Errorf("result = %+v, want upgraded to %s", res, tag)
	}
	if got, err := os.ReadFile(execPath); err != nil || !bytes.Equal(got, selfBytes(t)) {
		t.Errorf("exec binary not replaced (err=%v)", err)
	}
	assertInstallDirClean(t, dir, execPath)
	site.mu.Lock()
	defer site.mu.Unlock()
	if len(site.dlHits) != 2 {
		t.Errorf("downloaded assets = %v, want archive + checksums.txt", site.dlHits)
	}
}

// .old 遗留清理在 Update 分发层生效: go install 委托路径启动时同样清理
func TestUpdateGoInstallCleansStaleOldBinary(t *testing.T) {
	withRecordingGoShim(t, "github.com/wujunwei928/dev v0.3.0", 0)
	t.Setenv(fakeVersionEnv, "v0.3.0")

	dir := t.TempDir()
	execPath := filepath.Join(dir, binaryName())
	copySelf(t, execPath)
	stale := filepath.Join(dir, binaryName()+".old")
	if err := os.WriteFile(stale, []byte("stale-leftover"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := Update(UpdateOptions{
		Current:  VersionInfo{Method: MethodGoInstall, Version: "v0.2.0"},
		APIBase:  deadDownloadServer(t),
		ExecPath: execPath,
	}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale .old should be cleaned on go install path, stat err: %v", err)
	}
	if _, err := os.Stat(execPath); err != nil {
		t.Errorf("active binary should stay in place, stat err: %v", err)
	}
}
