package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func NewCmdEncode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode",
		Short: "字符串加密",
		Long:  "字符串加密",
	}

	// 子命令
	cmd.AddCommand(NewCmdEncodeMd5())
	cmd.AddCommand(NewCmdEncodeSha1())
	cmd.AddCommand(NewCmdEncodeBase64())
	cmd.AddCommand(NewCmdEncodeUrl())
	cmd.AddCommand(NewCmdEncodeUnicode())

	return cmd
}

func NewCmdEncodeMd5() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "md5",
		Short: "md5加密",
		Long:  "md5加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			hash := md5.Sum([]byte(encodeStr))
			content := hex.EncodeToString(hash[:])

			log.Printf("加密结果: %s\n", content)
		},
	}

	return cmd
}

func NewCmdEncodeSha1() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sha1",
		Short: "sha1加密",
		Long:  "sha1加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			hash := sha1.New()
			hash.Write([]byte(encodeStr))
			content := hex.EncodeToString(hash.Sum(nil))

			log.Printf("加密结果: %s", content)
		},
	}

	return cmd
}

func NewCmdEncodeBase64() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base64",
		Short: "base64加密",
		Long:  "base64加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			content := base64.StdEncoding.EncodeToString([]byte(encodeStr))

			log.Printf("加密结果: %s", content)
		},
	}

	return cmd
}

func NewCmdEncodeUrl() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url",
		Short: "url加密",
		Long:  "url加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			// 注意: 如果字符串中含有&, 需要用双引号 "a=1&b=2", 否则&后的字符会丢失
			content := url.QueryEscape(encodeStr)

			log.Printf("加密结果: %s", content)
		},
	}

	return cmd
}

func NewCmdEncodeUnicode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unicode",
		Short: "unicode加密",
		Long:  "unicode加密",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			encodeStr := args[0]

			content := strconv.QuoteToASCII(encodeStr)

			log.Printf("加密结果: %s", content)
		},
	}

	return cmd
}
