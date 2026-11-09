package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// re-exec 冒烟夹具: 子进程以测试二进制副本运行, 检出该环境变量时模拟 dev 的 --version 应答。
// 环境变量由具体测试通过 t.Setenv 设置, 随进程环境继承到子进程
const fakeVersionEnv = "DEV_UPDATE_TEST_FAKE_VERSION"

func TestMain(m *testing.M) {
	if v := os.Getenv(fakeVersionEnv); v != "" {
		fmt.Printf("dev version %s\n", v)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// selfPath 当前测试二进制自身路径
func selfPath(t *testing.T) string {
	t.Helper()
	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// copySelf 将测试二进制复制到 destPath, 供冒烟与交换测试当作「新版本二进制」使用
func copySelf(t *testing.T, destPath string) {
	t.Helper()
	src, err := os.Open(selfPath(t)) // #nosec G304 -- 测试内固定来源
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = src.Close() }()
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		t.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
}

// fakeBinaryPath 生成一个可执行的新版二进制副本 (含平台正确的文件名), 返回其路径
func fakeBinaryPath(t *testing.T, dir string) string {
	t.Helper()
	name := "dev"
	if runtime.GOOS == "windows" {
		name = "dev.exe"
	}
	p := filepath.Join(dir, name)
	copySelf(t, p)
	return p
}

// 冒烟测试通过: 输出版本与目标一致
func TestSmokeTestAcceptsMatchingVersion(t *testing.T) {
	bin := fakeBinaryPath(t, t.TempDir())
	t.Setenv(fakeVersionEnv, "v0.2.1")

	if err := smokeTest(bin, "v0.2.1"); err != nil {
		t.Fatalf("smokeTest should accept matching version, got: %v", err)
	}
}

// 冒烟测试通过: 目标无 v 前缀 (ldflags 注入形态) 而输出带 v 前缀时仍匹配
func TestSmokeTestNormalizesVPrefix(t *testing.T) {
	bin := fakeBinaryPath(t, t.TempDir())
	t.Setenv(fakeVersionEnv, "v0.2.1")

	if err := smokeTest(bin, "0.2.1"); err != nil {
		t.Fatalf("smokeTest should normalize v prefix, got: %v", err)
	}
}

// 冒烟测试失败: 输出版本与目标不符时拒绝替换
func TestSmokeTestRejectsWrongVersion(t *testing.T) {
	bin := fakeBinaryPath(t, t.TempDir())
	t.Setenv(fakeVersionEnv, "v0.9.9")

	err := smokeTest(bin, "v0.2.1")
	if err == nil {
		t.Fatal("smokeTest should reject mismatched version, got nil")
	}
	if !strings.Contains(err.Error(), "冒烟") {
		t.Errorf("error should mention smoke test, got: %v", err)
	}
	if !strings.Contains(err.Error(), "v0.2.1") {
		t.Errorf("error should name the target version, got: %v", err)
	}
}

// 冒烟测试失败: 二进制无法执行 (如损坏文件) 时报错
func TestSmokeTestRejectsUnrunnableBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows 对任意内容文件 CreateProcess 都会失败, 同样覆盖本场景
		_ = runtime.GOOS
	}
	dir := t.TempDir()
	name := "dev"
	if runtime.GOOS == "windows" {
		name = "dev.exe"
	}
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte("not a real executable"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := smokeTest(bin, "v0.2.1")
	if err == nil {
		t.Fatal("smokeTest should fail when binary cannot run, got nil")
	}
	if !strings.Contains(err.Error(), "冒烟") {
		t.Errorf("error should mention smoke test, got: %v", err)
	}
}

// 冒烟测试失败: 目标二进制缺失时报错
func TestSmokeTestRejectsMissingBinary(t *testing.T) {
	err := smokeTest(filepath.Join(t.TempDir(), "no-such-dev"), "v0.2.1")
	if err == nil {
		t.Fatal("smokeTest should fail when binary is missing, got nil")
	}
}

// --version 输出解析: 任一空白分隔 token 归一化后等于目标即匹配
func TestOutputMatchesVersion(t *testing.T) {
	cases := []struct {
		name   string
		output string
		target string
		want   bool
	}{
		{"cobra 默认输出", "dev version 0.2.1\n", "v0.2.1", true},
		{"目标无 v 前缀", "dev version v0.2.1", "0.2.1", true},
		{"多行输出中匹配", "dev version 0.2.1\nbuild go1.26\n", "v0.2.1", true},
		{"版本不符", "dev version 0.1.3\n", "v0.2.1", false},
		{"devel 输出", "dev version devel\n", "v0.2.1", false},
		{"空输出", "", "v0.2.1", false},
		{"复合 token 不误匹配", "dev-0.2.1\n", "0.2.1", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := outputMatchesVersion(c.output, c.target); got != c.want {
				t.Errorf("outputMatchesVersion(%q, %q) = %v, want %v", c.output, c.target, got, c.want)
			}
		})
	}
}
