package updater

import (
	"os"
	"path/filepath"
	"testing"
)

// writeSwapFixture 在 dir 下生成待交换的文件
func writeSwapFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// mustRead 断言文件存在并返回内容
func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// rename 交换: 新二进制顶替原位, 旧版改名备份后尽力删除
func TestSwapBinaryReplacesInPlace(t *testing.T) {
	dir := t.TempDir()
	execPath := writeSwapFixture(t, dir, "dev", "current-v1")
	newBin := writeSwapFixture(t, dir, "dev-new", "next-v2")

	if err := swapBinary(execPath, newBin); err != nil {
		t.Fatalf("swapBinary failed: %v", err)
	}
	if got := mustRead(t, execPath); got != "next-v2" {
		t.Errorf("exec binary content = %q, want %q", got, "next-v2")
	}
	if _, err := os.Stat(execPath + ".old"); !os.IsNotExist(err) {
		t.Errorf("backup .old should be cleaned after successful swap, stat err: %v", err)
	}
	if _, err := os.Stat(newBin); !os.IsNotExist(err) {
		t.Errorf("new binary should be moved away by rename, stat err: %v", err)
	}
}

// 替换失败时回滚备份, 旧版原位可用
func TestSwapBinaryRollsBackWhenReplaceFails(t *testing.T) {
	dir := t.TempDir()
	execPath := writeSwapFixture(t, dir, "dev", "current-v1")

	err := swapBinary(execPath, filepath.Join(dir, "missing-new"))
	if err == nil {
		t.Fatal("swapBinary should fail when new binary is missing, got nil")
	}
	if got := mustRead(t, execPath); got != "current-v1" {
		t.Errorf("exec binary content = %q, want rolled back %q", got, "current-v1")
	}
	if _, err := os.Stat(execPath + ".old"); !os.IsNotExist(err) {
		t.Errorf("backup .old should be restored (no leftover), stat err: %v", err)
	}
}

// 上次更新遗留的 .old 在下次更新启动时清理
func TestRemoveStaleOldBinary(t *testing.T) {
	dir := t.TempDir()
	execPath := writeSwapFixture(t, dir, "dev", "current-v1")
	stale := writeSwapFixture(t, dir, "dev.old", "leftover-old-v0")

	removeStaleOldBinary(execPath)

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale .old should be removed, stat err: %v", err)
	}
	if got := mustRead(t, execPath); got != "current-v1" {
		t.Errorf("exec binary content = %q, want untouched %q", got, "current-v1")
	}
}

// 无遗留 .old 时清理为空操作, 不报错
func TestRemoveStaleOldBinaryNoop(t *testing.T) {
	dir := t.TempDir()
	execPath := writeSwapFixture(t, dir, "dev", "current-v1")

	removeStaleOldBinary(execPath) // 不应 panic 或留下副作用

	if got := mustRead(t, execPath); got != "current-v1" {
		t.Errorf("exec binary content = %q, want untouched %q", got, "current-v1")
	}
}
