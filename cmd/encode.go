package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	EncodeTypeMd5     = "md5"     // md5加密
	EncodeTypeSha1    = "sha1"    // sha1加密
	EncodeTypeBase64  = "base64"  // base64加密
	EncodeTypeUrl     = "url"     // urlencode
	EncodeTypeUnicode = "unicode" // unicode
)

func NewCmdEncode() *cobra.Command {
	var encodeMode string

	// encodeCmd represents the encode command
	var encodeCmd = &cobra.Command{
		Use:   "encode",
		Short: "字符串加密",
		Long:  "字符串加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			var content string
			switch encodeMode {
			case EncodeTypeMd5:
				hash := md5.Sum([]byte(encodeStr))
				content = hex.EncodeToString(hash[:])
			case EncodeTypeSha1:
				hash := sha1.New()
				hash.Write([]byte(encodeStr))
				content = hex.EncodeToString(hash.Sum(nil))
			case EncodeTypeBase64:
				content = base64.StdEncoding.EncodeToString([]byte(encodeStr))
			case EncodeTypeUrl:
				// 注意: 如果字符串中含有&, 需要用双引号 "a=1&b=2", 否则&后的字符会丢失
				content = url.QueryEscape(encodeStr)
				//fmt.Println(encodeStr, content)
			case EncodeTypeUnicode:
				content = strconv.QuoteToASCII(encodeStr)
			default:
				log.Fatalf("暂不支持该加密模式, 请执行 help encode 查看帮助文档")
			}

			log.Printf("输出结果: %s", content)
		},
	}

	encodeCmd.Flags().StringVarP(&encodeMode, "mode", "m", "", strings.Join([]string{
		"请输入加密模式，支持模式如下：",
		EncodeTypeMd5 + "：md5加密",
		EncodeTypeSha1 + "：sha1加密",
		EncodeTypeBase64 + "：base64加密",
		EncodeTypeUrl + "：url加密",
		EncodeTypeUnicode + "：unicode加密",
	}, "\n"))

	return encodeCmd
}
