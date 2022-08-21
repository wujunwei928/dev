package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
)

var buildOutputName string

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use: "build [# times] [string]",

	Short: "交叉编译",
	Long:  `交叉编译, 打包各平台可执行文件`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New(pterm.Red("requires at least 1 arg(s), only received 0"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := filepath.Abs(args[0])

		osList := []string{"linux", "windows", "darwin"}
		archList := []string{"amd64", "arm64"}
		os.Mkdir("build_output", 0744)
		os.Chdir("build_output")
		for _, goOs := range osList {
			for _, goArch := range archList {
				outputName := buildOutputName
				if goOs == "windows" && !strings.Contains(outputName, ".exe") {
					outputName += ".exe"
				}

				os.Setenv("CGO_ENABLED", "0")
				os.Setenv("GOOS", goOs)
				os.Setenv("GOARCH", goArch)
				os.MkdirAll(filepath.Join(goOs, goArch), 0744)
				command := exec.Command("go", "build", "-o", path.Join(goOs, goArch, outputName), fileName)
				fmt.Println(command)
				err := command.Start()
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&buildOutputName, "name", "n", "", "output binary name")
}
