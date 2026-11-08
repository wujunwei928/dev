package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// verifyChecksum 校验下载产物文件的 sha256 与 checksums.txt 中同名条目一致。
// 任一不一致或条目缺失都返回错误, 由调用方中止安装 (旧版原位不动)
func verifyChecksum(archivePath, checksumsPath, archiveName string) error {
	content, err := os.ReadFile(checksumsPath)
	if err != nil {
		return fmt.Errorf("读取 checksums.txt 失败: %w", err)
	}
	want, ok := findChecksum(content, archiveName)
	if !ok {
		return fmt.Errorf("checksums.txt 中未找到 %s 的校验条目, 已拒绝安装", archiveName)
	}
	got, err := fileSHA256(archivePath)
	if err != nil {
		return fmt.Errorf("计算产物 sha256 失败: %w", err)
	}
	if got != want {
		return fmt.Errorf("sha256 校验失败: 期望 %s, 实际 %s, 已拒绝安装", want, got)
	}
	return nil
}

// findChecksum 从 checksums.txt 内容中查找 name 的 sha256 十六进制值。
// 行格式为 sha256sum 风格 "<hex64>  <name>", 容忍二进制标记 "*<name>" 与大小写差异
func findChecksum(content []byte, name string) (string, bool) {
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) != 2 {
			continue
		}
		fname := strings.TrimPrefix(fields[1], "*")
		if fname != name {
			continue
		}
		return strings.ToLower(fields[0]), true
	}
	return "", false
}

// fileSHA256 计算文件内容的 sha256 十六进制值
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 -- 路径来自 updater 自建的临时目录
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
