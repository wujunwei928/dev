package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// binaryName 当前平台发布产物内的二进制文件名
func binaryName() string {
	if runtime.GOOS == "windows" {
		return "dev.exe"
	}
	return "dev"
}

// extractBinary 从归档 (按文件后缀区分 .zip / .tar.gz) 中解出二进制到 destPath。
// 归档内容来自网络, 视为不可信输入: 只解出基名匹配的常规文件,
// 拒绝路径穿越 (zip slip) 条目, 跳过符号链接; 同名条目多处出现时取路径最浅者。
func extractBinary(archivePath, destPath string) error {
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractFromZip(archivePath, destPath)
	}
	return extractFromTarGz(archivePath, destPath)
}

// extractFromTarGz 解 tar.gz 归档中的二进制
func extractFromTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath) // #nosec G304 -- 路径来自 updater 自建的临时目录
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("打开归档失败: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	bestDepth := -1
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取归档失败: %w", err)
		}
		// 只接受常规文件: 符号链接/目录等一律跳过, 不解析链接目标
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != binaryName() {
			continue
		}
		depth, err := archiveEntryDepth(hdr.Name)
		if err != nil {
			continue // 逃逸条目直接忽略, 继续找安全候选
		}
		if bestDepth >= 0 && depth >= bestDepth {
			continue // 已有更浅 (或同层更早) 的候选
		}
		if err := writeBinary(destPath, tr); err != nil {
			return err
		}
		bestDepth = depth
	}
	if bestDepth < 0 {
		return fmt.Errorf("归档中未找到二进制 %s", binaryName())
	}
	return nil
}

// extractFromZip 解 zip 归档中的二进制
func extractFromZip(archivePath, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开归档失败: %w", err)
	}
	defer func() { _ = r.Close() }()

	bestDepth := -1
	for _, zf := range r.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(zf.Name) != binaryName() {
			continue
		}
		depth, err := archiveEntryDepth(zf.Name)
		if err != nil {
			continue // 逃逸条目直接忽略
		}
		if bestDepth >= 0 && depth >= bestDepth {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return fmt.Errorf("读取归档条目失败: %w", err)
		}
		err = writeBinary(destPath, rc)
		_ = rc.Close()
		if err != nil {
			return err
		}
		bestDepth = depth
	}
	if bestDepth < 0 {
		return fmt.Errorf("归档中未找到二进制 %s", binaryName())
	}
	return nil
}

// archiveEntryDepth 归一化归档条目路径并返回其深度; 含 .. 段或绝对路径时拒绝 (防 zip slip)
func archiveEntryDepth(name string) (int, error) {
	if strings.ContainsAny(name, "\x00") {
		return 0, fmt.Errorf("条目路径含非法字符: %q", name)
	}
	p := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(p) || strings.HasPrefix(p, `\`) || filepath.VolumeName(p) != "" {
		return 0, fmt.Errorf("条目路径为绝对路径: %q", name)
	}
	depth := 0
	for _, seg := range strings.Split(p, string(os.PathSeparator)) {
		switch seg {
		case "..":
			return 0, fmt.Errorf("条目路径逃逸归档根目录: %q", name)
		case ".", "":
			continue
		}
		depth++
	}
	if depth == 0 {
		return 0, fmt.Errorf("条目路径为空: %q", name)
	}
	return depth, nil
}

// writeBinary 将条目内容写入 destPath, 非 Windows 带可执行权限
func writeBinary(destPath string, r io.Reader) error {
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o666
	}
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode) // #nosec G304 -- 路径来自 updater 自建的暂存目录
	if err != nil {
		return fmt.Errorf("写入暂存二进制失败: %w", err)
	}
	if _, err := io.Copy(f, io.LimitReader(r, MaxDownloadBodySize+1)); err != nil {
		_ = f.Close()
		return fmt.Errorf("解出二进制失败: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("写入暂存二进制失败: %w", err)
	}
	return nil
}
