package cmd

import (
	"encoding/base64"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	DecodeTypeBase64  = "base64"  // base64解密
	DecodeTypeUrl     = "url"     // urldecode
	DecodeTypeUnicode = "unicode" // unicode
)

var decodeMode string

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "字符串解密",
	Long:  "字符串解密",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		decodeStr := args[0]

		var content string
		switch decodeMode {
		case DecodeTypeBase64:
			decodeByte, err := base64.StdEncoding.DecodeString(decodeStr)
			if err != nil {
				log.Fatalf("base64 decode fail: %s", err.Error())
			}
			content = string(decodeByte)
		case DecodeTypeUrl:
			s, err := url.QueryUnescape(decodeStr)
			if err != nil {
				log.Fatalf("url decode fail: %s", err.Error())
			}
			content = s
		case DecodeTypeUnicode:
			// unicode转换中文需要的格式 "内容" , 注意要传双引号
			if !strings.Contains(decodeStr, `"`) {
				decodeStr = `"` + decodeStr + `"`
			}
			s, err := strconv.Unquote(decodeStr)
			if err != nil {
				log.Fatalf("unicode decode fail: %s", err.Error())
			}
			content = s
		default:
			log.Fatalf("暂不支持该解密模式, 请执行 help decode 查看帮助文档")
		}

		log.Printf("输出结果: %s", content)
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().StringVarP(&decodeMode, "mode", "m", "", strings.Join([]string{
		"请输入解密模式, 支持模式如下: ",
		DecodeTypeBase64 + "：base64解密",
		DecodeTypeUrl + "：url解密",
		DecodeTypeUnicode + "：unicode解密",
	}, "\n"))
}
