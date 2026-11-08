package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wujunwei928/dev/internal/updater"
)

// 更新配置项
const (
	UpdateConfigAPIBase = "update.api_base" // GitHub API base, 默认 api.github.com
)

func NewCmdUpdate() *cobra.Command {
	var checkOnly bool
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "检查并安装更新",
		Long:  "检查并安装更新, --check 只查询最新版本不安装, --version <ver> 安装指定版本 (可降级回滚)",
		RunE: func(cmd *cobra.Command, args []string) error {
			vi := updater.CurrentVersion(Version)
			apiBase := viper.GetString(UpdateConfigAPIBase)
			out := cmd.OutOrStdout()

			if checkOnly {
				res, err := updater.Check(vi, apiBase)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "当前版本: %s\n", res.Current)
				fmt.Fprintf(out, "最新版本: %s\n", res.Latest)
				if res.UpToDate {
					fmt.Fprintln(out, "已是最新版本 (already up to date)")
				} else {
					fmt.Fprintf(out, "发现新版本 %s 可用\n", res.Latest)
				}
				return nil
			}

			if vi.Method == updater.MethodLocal {
				fmt.Fprintln(out, "警告: 当前为本地构建版本, 更新将替换为 GitHub Releases 的官方发布版")
			}

			res, err := updater.Update(updater.UpdateOptions{
				Current: vi,
				Target:  targetVersion,
				APIBase: apiBase,
			})
			if err != nil {
				return err
			}
			if res.UpToDate {
				fmt.Fprintln(out, "已是最新版本 (already up to date)")
				return nil
			}
			if res.GoMissingFallback {
				fmt.Fprintln(out, "未找到 go 命令, 已回退到 GitHub Releases 下载替换路径")
			}
			fmt.Fprintf(out, "已更新到 %s (原版本 %s)\n", res.To, res.From)
			if !res.ChecksumVerified && res.Method != updater.MethodGoInstall {
				fmt.Fprintln(out, "注意: 该版本未提供 checksums.txt, 未经 sha256 校验")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "只查询最新版本, 不安装更新")
	cmd.Flags().StringVar(&targetVersion, "version", "", "安装指定版本 (如 v0.2.0), 可用于降级回滚")
	return cmd
}
