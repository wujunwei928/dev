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

	cmd := &cobra.Command{
		Use:   "update",
		Short: "检查并安装更新",
		Long:  "检查并安装更新, --check 只查询最新版本不安装",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !checkOnly {
				// 安装动作由后续版本提供 (#8/#9), 本版仅立起命令骨架
				fmt.Fprintln(cmd.OutOrStdout(), "自动更新尚未提供, 当前仅支持 --check 查询最新版本")
				return nil
			}

			res, err := updater.Check(updater.CurrentVersion(Version), viper.GetString(UpdateConfigAPIBase))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "当前版本: %s\n", res.Current)
			fmt.Fprintf(out, "最新版本: %s\n", res.Latest)
			if res.UpToDate {
				fmt.Fprintln(out, "已是最新版本 (already up to date)")
			} else {
				fmt.Fprintf(out, "发现新版本 %s 可用\n", res.Latest)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "只查询最新版本, 不安装更新")
	return cmd
}
