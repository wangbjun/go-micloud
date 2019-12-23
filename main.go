package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
	"go-micloud/command"
	"go-micloud/lib/line"
	"io"
	"os"
	"strings"
)

func main() {
	app := &cli.App{
		Name:    "Go-MiCloud",
		Usage:   "MiCloud Third Party Client Written By Golang",
		Version: "1.0",
		Commands: []*cli.Command{
			command.Login(),
			command.List(),
			command.Download(),
			command.Cd(),
			command.Upload(),
			command.Share(),
		},
		CommandNotFound: func(c *cli.Context, command string) {
			fmt.Printf("Command \"%s\" not found\n", command)
		},
	}
	for {
		commandLine, err := line.CsLiner.Prompt()
		if err != nil {
			if err == liner.ErrPromptAborted || err == io.EOF {
				_ = line.CsLiner.Close()
				return
			}
			fmt.Printf("===> Prompt Error: %s\n", err)
			continue
		}
		line.CsLiner.AppendHistory(commandLine)
		var args = append(
			[]string{os.Args[0]},
			strings.Split(commandLine, " ")...)
		err = app.Run(args)
		if err != nil {
			fmt.Printf("===> Error: %s\n", err)
			continue
		}
	}
}
