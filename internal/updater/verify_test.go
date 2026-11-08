package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 期望值取自独立事实源 (sha256sum 预计算字面量), 不复用被测代码的哈希结果
const (
	archiveContent  = "dev binary payload v2"
	archiveSHA256   = "e798f4435668ba756f0883aed8d6bdc61c4d41f430c72a4a1e305a4716ef8a30"
	otherSHA256     = "3f99d178af0abef274d94baef9a2ef7108940420c100372d78d356f6610d7f4e"
	checksumsFakeNm = "dev_0.2.1_windows_amd64.zip"
)

// writeArchive 写出待校验的产物文件, 返回其路径
func writeArchive(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, checksumsFakeNm)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// sha256 与 checksums.txt 记录一致时通过
func TestVerifyChecksumAcceptsMatchingArchive(t *testing.T) {
	dir := t.TempDir()
	archivePath := writeArchive(t, dir, archiveContent)
	checksumsPath := filepath.Join(dir, "checksums.txt")
	content := archiveSHA256 + "  " + checksumsFakeNm + "\n"
	if err := os.WriteFile(checksumsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := verifyChecksum(archivePath, checksumsPath, checksumsFakeNm); err != nil {
		t.Fatalf("verifyChecksum should accept matching archive, got: %v", err)
	}
}

// sha256 不匹配时拒绝安装
func TestVerifyChecksumRejectsMismatch(t *testing.T) {
	dir := t.TempDir()
	archivePath := writeArchive(t, dir, archiveContent)
	checksumsPath := filepath.Join(dir, "checksums.txt")
	content := otherSHA256 + "  " + checksumsFakeNm + "\n"
	if err := os.WriteFile(checksumsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := verifyChecksum(archivePath, checksumsPath, checksumsFakeNm)
	if err == nil {
		t.Fatal("verifyChecksum should reject mismatched archive, got nil")
	}
	if !strings.Contains(err.Error(), "sha256") {
		t.Errorf("error should mention sha256, got: %v", err)
	}
}

// checksums.txt 中没有该产物条目时拒绝安装, 不放行
func TestVerifyChecksumRejectsMissingEntry(t *testing.T) {
	dir := t.TempDir()
	archivePath := writeArchive(t, dir, archiveContent)
	checksumsPath := filepath.Join(dir, "checksums.txt")
	content := archiveSHA256 + "  some-other-file.zip\n"
	if err := os.WriteFile(checksumsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	err := verifyChecksum(archivePath, checksumsPath, checksumsFakeNm)
	if err == nil {
		t.Fatal("verifyChecksum should reject archive missing from checksums.txt, got nil")
	}
	if !strings.Contains(err.Error(), checksumsFakeNm) {
		t.Errorf("error should name the missing asset, got: %v", err)
	}
}

// 大写十六进制与二进制标记 (*name) 形态的条目同样可匹配
func TestVerifyChecksumAcceptsUppercaseAndBinaryMarker(t *testing.T) {
	dir := t.TempDir()
	archivePath := writeArchive(t, dir, archiveContent)
	checksumsPath := filepath.Join(dir, "checksums.txt")
	content := strings.ToUpper(archiveSHA256) + " *" + checksumsFakeNm + "\n"
	if err := os.WriteFile(checksumsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := verifyChecksum(archivePath, checksumsPath, checksumsFakeNm); err != nil {
		t.Fatalf("verifyChecksum should accept uppercase hash with binary marker, got: %v", err)
	}
}
