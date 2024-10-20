package cmd

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version = "0.1.3" // 版本号

	cfgFile string // 自定义配置路径, 类似 ~/.bashrc
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dev",
	Short:   "研发工具箱",
	Long:    `研发工具箱`,
	Version: Version, // 指定版本号: 会有 -v 和 --version 选项, 用于打印版本号
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 加载默认配置
	cobra.OnInitialize(initConfig)
	// 指定自定义配置文件
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dev.yaml)")

	// 添加子命令
	rootCmd.AddCommand(NewCmdConsole()) // 类似ipython的交互式命令行
	rootCmd.AddCommand(NewCmdDecode())  // 字符串解密
	rootCmd.AddCommand(NewCmdEncode())  // 字符串加密
	rootCmd.AddCommand(NewCmdInstall()) // 安装命令到PATH
	rootCmd.AddCommand(NewCmdSearch())  // 搜索
	rootCmd.AddCommand(NewCmdOpen())    // 打开文件或目录
	rootCmd.AddCommand(NewCmdTime())    // 时间转换
	rootCmd.AddCommand(NewCmdHttp())    // http服务
	rootCmd.AddCommand(NewCmdWord())    // 单词格式转换
	rootCmd.AddCommand(NewCmdUrl())     // 打开网址，文件夹或文件
	rootCmd.AddCommand(NewCmdSql())     // sql相关
	rootCmd.AddCommand(NewCmdGo())      // go相关
}

func initConfig() {
	// Use config file from the flag.
	if len(cfgFile) > 0 {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("viper read config fail: %s", err.Error())
		}
	}

	// use config from home directory!
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("get home dir fail: %s", err.Error())
	}

	// Search config in home directory with name ".dev.yaml"
	defaultCfgFile, err := filepath.Abs(path.Join(home, ".dev.yaml"))
	if err != nil {
		log.Fatalf("get default config path fail: %s", err.Error())
	}
	viper.SetConfigFile(defaultCfgFile)
	if _, err := os.Stat(defaultCfgFile); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("check default config file is exists fail: %s", err.Error())
		}

		// 用户未指定config路径时, 如果默认配置文件不存在, 自动创建
		ViperInitSet()
		err := viper.SafeWriteConfigAs(defaultCfgFile)
		if err != nil {
			log.Fatalf("default config not exist, auto create fail: %s", err.Error())
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("viper read config fail: %s", err.Error())
	}
}

// ViperInitSet viper 初始化设置
func ViperInitSet() {
	// http
	viper.Set(HttpConfigPort, DefaultHttpPort)

	// search
	viper.Set(SearchConfigEngine, DefaultSearchEngine)
	viper.Set(SearchConfigType, DefaultSearchType)
	viper.Set(SearchConfigCliIsDesc, DefaultCliIsDesc)

	// sql
	viper.Set(SqlConfigType, DefaultSqlType)
	viper.Set(SqlConfigHost, DefaultSqlHost)
	viper.Set(SqlConfigUserName, "")
	viper.Set(SqlConfigPassword, "")
	viper.Set(SqlConfigDb, "")
	viper.Set(SqlConfigCharset, DefaultSqlCharset)
}
