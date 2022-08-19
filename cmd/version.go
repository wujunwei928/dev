package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version = "0.1.1"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本号",
	Long:  `打印版本号`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dev " + Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
