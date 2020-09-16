package zlog

import (
	"fmt"
	"go-micloud/pkg/color"
	"time"
)

func Info(msg string) {
	fmt.Printf(color.Green(time.Now().Format("2006-01-02 15:04:05")+" ### 提示: %s\n"), msg)
	Logger.Info(msg)
}

func Error(msg string) {
	fmt.Printf(color.Red(time.Now().Format("2006-01-02 15:04:05")+" ### 错误: %s\n"), msg)
	Logger.Error(msg)
}
