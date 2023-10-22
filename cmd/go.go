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
		pkgName  string
		isUpdate bool

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
			"github.com/xuri/excelize",
		}
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "模糊检索常用go module包",
		Long:  "模糊检索常用go module包",
		RunE: func(cmd *cobra.Command, args []string) error {

			searchPkgList := lo.Filter(commonUsePkgList, func(item string, index int) bool {
				return strings.Contains(strings.ToLower(item), strings.ToLower(pkgName))
			})

			if len(searchPkgList) == 0 {
				log.Fatal("未找到相关包")
			}

			// 打印检索到的包
			printPrefix := "go get"
			if isUpdate {
				printPrefix = "go get -u"
			}
			fmt.Println("检索到的包:")
			for _, pkg := range searchPkgList {
				fmt.Println(printPrefix, pkg)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgName, "pkg_name", "p", "", "包名关键字")
	cmd.Flags().BoolVarP(&isUpdate, "update", "u", false, "是否更新")

	return cmd
}

func (c GoSubCmd) NewCmdBuild() *cobra.Command {
	var buildOutputName string

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

	cmd.Flags().StringVarP(&buildOutputName, "name", "n", "", "output binary name")

	return cmd
}
