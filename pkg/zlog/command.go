package zlog

import (
	"fmt"
	"go-micloud/pkg/color"
)

func Info(msg string) {
	fmt.Printf(color.Green("### Info: %s\n"), msg)
}

func Error(msg string) {
	fmt.Printf(color.Red("### Error: %s\n"), msg)
}
