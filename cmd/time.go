package cmd

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/wujunwei928/dev/internal/timer"

	"github.com/spf13/cobra"
)

// 解析时间抽参数
var parseTimeStamp int64

// 计算时间参数
var calculateTime string
var calculateDuration string

// timeCmd represents the time command
var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "时间格式处理",
	Long:  `时间格式处理`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

var parseTimeCmd = &cobra.Command{
	Use:   "parse",
	Short: "解析时间戳为标准格式, 不传默认当前时间",
	Long:  "解析时间戳为标准格式, 不传默认当前时间",
	Run: func(cmd *cobra.Command, args []string) {
		nowTime := time.Unix(parseTimeStamp, 0)
		log.Printf("当前时间: %s , %d", nowTime.Format("2006-01-02 15:04:05"), nowTime.Unix())
	},
}

var calculateTimeCmd = &cobra.Command{
	Use:   "calc",
	Short: "计算所需时间",
	Long:  "计算所需时间",
	Run: func(cmd *cobra.Command, args []string) {
		var currentTimer time.Time
		var layout = "2006-01-02 15:04:05"
		if calculateTime == "" {
			currentTimer = time.Now()
		} else {
			var err error
			space := strings.Count(calculateTime, " ")
			if space == 0 {
				layout = "2006-01-02"
			}
			currentTimer, err = time.Parse(layout, calculateTime)
			if err != nil {
				t, _ := strconv.Atoi(calculateTime)
				currentTimer = time.Unix(int64(t), 0)
			}
		}
		t, err := timer.GetCalculateTime(currentTimer, calculateDuration)
		if err != nil {
			log.Fatalf("timer.GetCalculateTime err: %v", err)
		}

		log.Printf("输出结果: %s , %d", t.Format(layout), t.Unix())
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)

	timeCmd.AddCommand(parseTimeCmd)
	parseTimeCmd.Flags().Int64VarP(&parseTimeStamp, "timestamp", "t", time.Now().Unix(), "解析时间戳, 不传默认当前时间戳")
	timeCmd.AddCommand(calculateTimeCmd)
	calculateTimeCmd.Flags().StringVarP(&calculateTime, "calculate", "c", "", ` 需要计算的时间，有效单位为时间戳或已格式化后的时间 `)
	calculateTimeCmd.Flags().StringVarP(&calculateDuration, "calculateDuration", "d", "", ` 持续时间，有效时间单位为"ns", "us" (or "µ s"), "ms", "s", "m", "h"`)
}
