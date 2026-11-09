package updater

import (
	"fmt"
	"os"
)

// oldBinarySuffix 旧版本二进制备份后缀。
// Windows 上运行中的 exe 可 rename 不可删: 更新通过 rename 交换完成, 运行中的旧版改名为 .old,
// 删除失败不阻塞更新, 留待下次 update 启动时清理
const oldBinarySuffix = ".old"

// swapBinary 原子交换: 现役二进制改名 .old 备份, 新二进制顶替原位。
// 顶替失败时回滚备份, 保证旧版始终原位可用
func swapBinary(execPath, newBinaryPath string) error {
	oldPath := execPath + oldBinarySuffix
	_ = os.Remove(oldPath) // Update 启动时已清过, 此处兜底

	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("备份旧版本失败: %w", err)
	}
	if err := os.Rename(newBinaryPath, execPath); err != nil {
		if rbErr := os.Rename(oldPath, execPath); rbErr != nil {
			return fmt.Errorf("替换新版本失败: %v; 回滚旧版也失败: %v (旧版备份保留在 %s)", err, rbErr, oldPath)
		}
		return fmt.Errorf("替换新版本失败: %w (已回滚, 旧版原位可用)", err)
	}
	// 尽力清理备份: 非 Windows 直接删除; Windows 上备份若是运行中的旧映像会失败,
	// 静默留给下次 update 启动清理
	_ = os.Remove(oldPath)
	return nil
}

// removeStaleOldBinary 清理上次更新遗留的 .old 备份, 尽力而为 (失败不阻塞本次更新)
func removeStaleOldBinary(execPath string) {
	_ = os.Remove(execPath + oldBinarySuffix)
}
