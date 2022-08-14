package cmd

import (
	"encoding/base64"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

const (
	DecodeTypeBase64 = "base64" // base64解密
	DecodeTypeUrl    = "url"    // urldecode
)

var decodeDesc = strings.Join([]string{
	"该子命令支持各种解密，模式如下：",
	DecodeTypeBase64 + "：base64解密",
	DecodeTypeUrl + "：url解密",
}, "\n")

var decodeStr string
var decodeMode string

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "字符串解密",
	Long:  decodeDesc,
	Run: func(cmd *cobra.Command, args []string) {
		var content string
		switch decodeMode {
		case DecodeTypeBase64:
			decodeByte, err := base64.StdEncoding.DecodeString(decodeStr)
			if err != nil {
				log.Fatalf("base64 decode fail: %s", err.Error())
			}
			content = string(decodeByte)
		case DecodeTypeUrl:
			decodeStr, err := url.QueryUnescape(decodeStr)
			if err != nil {
				log.Fatalf("url decode fail: %s", err.Error())
			}
			content = decodeStr
		default:
			log.Fatalf("暂不支持该解密模式, 请执行 help decode 查看帮助文档")
		}

		log.Printf("输出结果: %s", content)
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().StringVarP(&decodeStr, "str", "s", "", "请输入单词内容")
	decodeCmd.Flags().StringVarP(&decodeMode, "mode", "m", "", "请输入加密模式")
}
