package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const payloadA = "fake-dev-binary A\n"
const payloadB = "fake-dev-binary B\n"

// binaryFixtureName 当前平台归档内的二进制名
func binaryFixtureName() string {
	if runtime.GOOS == "windows" {
		return "dev.exe"
	}
	return "dev"
}

// buildTarGz 将条目打包为 tar.gz 字节流; nil Content 表示目录项, kind 覆盖条目类型 (0 = 常规文件)
func buildTarGz(t *testing.T, entries []struct {
	Name  string
	Kind  byte
	Link  string
	Bytes []byte
}) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.Name, Mode: 0o755}
		switch {
		case e.Kind == tar.TypeSymlink:
			hdr.Typeflag = tar.TypeSymlink
			hdr.Linkname = e.Link
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
		default:
			hdr.Typeflag = tar.TypeReg
			hdr.Size = int64(len(e.Bytes))
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(e.Bytes); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildZip 将条目打包为 zip 字节流
func buildZip(t *testing.T, entries []struct {
	Name  string
	Bytes []byte
}) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, err := zw.Create(e.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(e.Bytes); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// writeArchiveFile 归档字节流落盘为指定后缀文件, 返回路径
func writeArchiveFile(t *testing.T, data []byte, suffix string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "asset"+suffix)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// tar.gz 归档 (历史 Release 产物形态): 解出二进制, 忽略其他文件
func TestExtractBinaryFromTarGz(t *testing.T) {
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "README.md", Bytes: []byte("docs")},
		{Name: bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payloadA {
		t.Errorf("extracted content = %q, want %q", got, payloadA)
	}
}

// tar.gz 归档内二进制位于子目录时同样解出
func TestExtractBinaryFromTarGzNested(t *testing.T) {
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "packaged/" + bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != payloadA {
		t.Errorf("extracted content = %q, want %q", got, payloadA)
	}
}

// 同名二进制多处出现时取路径最浅者 (根级优先)
func TestExtractBinaryPrefersShallowestMatch(t *testing.T) {
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "nested/" + bin, Bytes: []byte(payloadB)},
		{Name: bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != payloadA {
		t.Errorf("extracted content = %q, want shallowest root entry %q", got, payloadA)
	}
}

// zip 归档 (goreleaser Windows 产物形态): 解出二进制
func TestExtractBinaryFromZip(t *testing.T) {
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildZip(t, []struct {
		Name  string
		Bytes []byte
	}{
		{Name: "LICENSE", Bytes: []byte("MIT")},
		{Name: bin, Bytes: []byte(payloadA)},
	}), ".zip")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != payloadA {
		t.Errorf("extracted content = %q, want %q", got, payloadA)
	}
}

// 归档条目路径穿越 (zip slip) 必须拒绝, 且不写出目标位置之外的文件
func TestExtractBinaryRejectsPathTraversal(t *testing.T) {
	outDir := t.TempDir()
	parent := filepath.Dir(outDir)
	bin := binaryFixtureName()
	escaped := filepath.Join(parent, bin)

	// zip: ../<bin> 基名与二进制一致但路径逃逸
	zipPath := writeArchiveFile(t, buildZip(t, []struct {
		Name  string
		Bytes []byte
	}{
		{Name: "../" + bin, Bytes: []byte(payloadA)},
	}), ".zip")
	err := extractBinary(zipPath, filepath.Join(outDir, bin))
	if err == nil {
		t.Fatal("extractBinary should reject traversal entry, got nil")
	}
	if _, statErr := os.Stat(escaped); statErr == nil {
		t.Errorf("traversal entry must not be written outside destination, found %s", escaped)
	}

	// tar.gz: 同样的逃逸形态
	tarPath := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "../" + bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")
	err = extractBinary(tarPath, filepath.Join(outDir, bin))
	if err == nil {
		t.Fatal("extractBinary should reject traversal entry, got nil")
	}
	if _, statErr := os.Stat(escaped); statErr == nil {
		t.Errorf("traversal entry must not be written outside destination, found %s", escaped)
	}
}

// tar 符号链接条目不追写、不解析, 防止借链接落盘任意文件
func TestExtractBinarySkipsSymlinkEntries(t *testing.T) {
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: bin, Kind: tar.TypeSymlink, Link: "../../outside-target"},
		{Name: "real/" + bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != payloadA {
		t.Errorf("extracted content = %q, want regular entry %q (symlink must be skipped)", got, payloadA)
	}
}

// 归档中不含二进制时报错
func TestExtractBinaryMissingBinaryFails(t *testing.T) {
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: "README.md", Bytes: []byte("docs")},
	}), ".tar.gz")

	err := extractBinary(archive, filepath.Join(t.TempDir(), binaryFixtureName()))
	if err == nil {
		t.Fatal("extractBinary should fail when archive has no binary, got nil")
	}
	if !strings.Contains(err.Error(), binaryFixtureName()) {
		t.Errorf("error should name the expected binary, got: %v", err)
	}
}

// 非 Windows 平台解出的二进制应带可执行权限 (Windows 无该语义)
func TestExtractBinarySetsExecutableMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows 无 POSIX 执行位")
	}
	bin := binaryFixtureName()
	archive := writeArchiveFile(t, buildTarGz(t, []struct {
		Name  string
		Kind  byte
		Link  string
		Bytes []byte
	}{
		{Name: bin, Bytes: []byte(payloadA)},
	}), ".tar.gz")

	dest := filepath.Join(t.TempDir(), bin)
	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary failed: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("extracted binary mode = %v, want owner-executable", info.Mode().Perm())
	}
}
