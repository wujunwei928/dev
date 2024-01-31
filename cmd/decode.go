package cmd

import (
	"encoding/base64"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type decodeSubCmd struct{}

func NewCmdDecode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "字符串解密",
		Long:  "字符串解密",
	}

	subCmd := decodeSubCmd{}
	cmd.AddCommand(subCmd.NewCmdBase64())
	cmd.AddCommand(subCmd.NewCmdUrl())
	cmd.AddCommand(subCmd.NewCmdUnicode())

	return cmd
}

func (d decodeSubCmd) NewCmdBase64() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base64",
		Short: "base64解密",
		Long:  "base64解密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			decodeStr := args[0]
			decodeByte, err := base64.StdEncoding.DecodeString(decodeStr)
			if err != nil {
				log.Fatalf("base64 decode fail: %s", err.Error())
			}
			content := string(decodeByte)

			log.Printf("输出结果: %s", content)
		},
	}

	return cmd
}

func (d decodeSubCmd) NewCmdUrl() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url",
		Short: "url解密",
		Long:  "url解密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			decodeStr := args[0]
			content, err := url.QueryUnescape(decodeStr)
			if err != nil {
				log.Fatalf("url decode fail: %s", err.Error())
			}

			log.Printf("输出结果: %s", content)
		},
	}

	return cmd
}

func (d decodeSubCmd) NewCmdUnicode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unicode",
		Short: "unicode解密",
		Long:  "unicode解密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			decodeStr := args[0]
			// unicode转换中文需要的格式 "内容" , 注意要传双引号
			if !strings.Contains(decodeStr, `"`) {
				decodeStr = `"` + decodeStr + `"`
			}
			content, err := strconv.Unquote(decodeStr)
			if err != nil {
				log.Fatalf("unicode decode fail: %s", err.Error())
			}
			log.Printf("输出结果: %s", content)
		},
	}

	return cmd
}
