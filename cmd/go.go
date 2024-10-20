package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type GoSubCmd struct{}

func NewCmdGo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go",
		Short: "go相关二级命令",
		Long:  "go相关二级命令",
	}

	// 添加子命令
	goSubCmd := &GoSubCmd{}
	cmd.AddCommand(goSubCmd.newSearchCmd())
	cmd.AddCommand(goSubCmd.NewCmdBuild())

	return cmd
}

func (c GoSubCmd) newSearchCmd() *cobra.Command {
	var (
		pkgName   string
		isInstall bool
		isUpdate  bool

		commonUsePkgList = []string{
			"github.com/hibiken/asynq",
			"github.com/spf13/cobra",
			"github.com/spf13/viper",
			"github.com/spf13/cast",
			"github.com/zeromicro/go-zero",
			"github.com/gin-gonic/gin",
			"gorm.io/gorm",
			"gorm.io/driver/mysql",
			"github.com/glebarez/sqlite",
			"github.com/google/uuid",
			"github.com/robfig/cron/v3",
			"github.com/samber/lo",
			"github.com/tidwall/gjson",
			"github.com/tidwall/sjson",
			"google.golang.org/grpc",
			"github.com/gocolly/colly/v2",
			"github.com/PuerkitoBio/goquery",
			"github.com/segmentio/kafka-go",
			"github.com/redis/go-redis/v9",
			"go.mongodb.org/mongo-driver/mongo",
			"github.com/elastic/go-elasticsearch/v8",
			"github.com/golang-module/carbon/v2",
			"github.com/xuri/excelize/v2",
			"github.com/pterm/pterm",
		}
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "模糊检索常用go module包",
		Long:  "模糊检索常用go module包",
		RunE: func(cmd *cobra.Command, args []string) error {
			filterPkgList := commonUsePkgList
			if len(pkgName) > 0 {
				filterPkgList = lo.Filter(commonUsePkgList, func(item string, index int) bool {
					return strings.Contains(strings.ToLower(item), strings.ToLower(pkgName))
				})
			}

			if len(filterPkgList) == 0 {
				log.Fatal("未找到相关包")
			}

			searchPkgList := filterPkgList
			if len(filterPkgList) > 10 {
				options := make([]string, 0, len(filterPkgList))
				for _, pkg := range filterPkgList {
					options = append(options, pkg)
				}
				searchPkgList, _ = pterm.DefaultInteractiveMultiselect.
					WithOptions(options).
					WithDefaultText("请选择需要的go package: [可上方向键移动, 或输入关键字符检索, 回车选择]").
					WithMaxHeight(6).
					Show()
			}

			// 打印检索到的包
			printPrefix := []string{"go", "get"}
			if isUpdate {
				printPrefix = append(printPrefix, "-u")
			}
			fmt.Println("检索到的包:")
			for _, pkg := range searchPkgList {
				if isInstall {
					command := exec.Command("go", "get", pkg)
					if isUpdate {
						command = exec.Command("go", "get", "-u", pkg)
					}
					installOutput, err := command.CombinedOutput()
					if err != nil {
						log.Fatalln("安装失败", err.Error())
					} else {
						fmt.Println(string(installOutput))
					}
				} else {
					fmt.Println(strings.Join(append(printPrefix, pkg), " "))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&isInstall, "install", "i", false, "是否安装")
	cmd.Flags().BoolVarP(&isUpdate, "update", "u", false, "是否更新, 与install同时使用")
	cmd.Flags().StringVarP(&pkgName, "pkg_name", "p", "", "包名关键字")

	return cmd
}

func (c GoSubCmd) NewCmdBuild() *cobra.Command {
	var (
		err             error
		buildOutputName string
		buildOutputPath = "build_output"
	)

	cmd := &cobra.Command{
		Use:   "build [# times] [string]",
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
			os.Mkdir(buildOutputPath, 0744)
			os.Chdir(buildOutputPath)
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
					_, err = command.CombinedOutput()
					if err != nil {
						log.Fatalln("start build fail", err.Error())
					}
				}
			}
		},
	}

	cmd.Flags().StringVarP(&buildOutputName, "name", "n", "", "output binary name")
	cmd.Flags().StringVarP(&buildOutputPath, "path", "p", buildOutputPath, "output binary save path")

	return cmd
}
