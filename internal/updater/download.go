package updater

import (
	"fmt"
	"io"
	"os"
)

// downloadFile 将 url 内容流式下载到 destPath, 超过 MaxDownloadBodySize 时中止。
// 走独立的下载客户端 (遵循系统代理、限制跳转), 响应体不整体驻留内存
func downloadFile(url, destPath string) error {
	resp, err := GetDownloadClient().R().SetDoNotParseResponse(true).Get(url)
	if err != nil {
		return fmt.Errorf("下载失败, 请检查网络: %w", err)
	}
	raw := resp.RawResponse
	if raw == nil || raw.Body == nil {
		return fmt.Errorf("下载失败: 响应缺少内容")
	}
	defer func() { _ = raw.Body.Close() }()
	if !resp.IsSuccess() {
		return fmt.Errorf("下载失败: HTTP %s", resp.Status())
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600) // #nosec G304 -- 路径来自 updater 自建的临时目录
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	n, err := io.Copy(f, io.LimitReader(raw.Body, MaxDownloadBodySize+1))
	closeErr := f.Close()
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("写入临时文件失败: %w", closeErr)
	}
	if n > MaxDownloadBodySize {
		return fmt.Errorf("下载中止: 产物超过 %dMB 大小上限", MaxDownloadBodySize>>20)
	}
	return nil
}
