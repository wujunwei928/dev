package cmd

import (
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	cfgFile string // 自定义配置路径, 类似 ~/.bashrc
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bd",
	Short: "研发工具箱",
	Long:  `研发工具箱`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bd.yaml)")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("get home dir fail: %s", err.Error())
		}

		// Search config in home directory with name ".bd.yaml"
		viper.AddConfigPath(home)
		viper.SetConfigName(".bd")
		viper.SetConfigType("yaml")

	}

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			if len(cfgFile) > 0 {
				// 用户指定自定义配置路径时, 正常报文件未找到错误
				log.Fatalf("viper read config fail: %s", err.Error())
			}
			// 用户未指定config路径时, 如果默认配置不存在, 自动创建
			viper.Set("default_search_engine", "kaifa")
			err := viper.SafeWriteConfig()
			if err != nil {
				log.Fatalf("default config not exist, auto create fail: %s", err.Error())
			}
		default:
			log.Fatalf("viper read config fail: %s", err.Error())
		}
	}
}
