package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/command"
	"go-micloud/internal/file"
	"go-micloud/internal/folder"
	"go-micloud/internal/user"
	"go-micloud/pkg/line"
	"go-micloud/pkg/zlog"
	"io"
	"os"
	"strings"
)

func main() {
	httpApi := file.NewApi(user.NewUser())
	if httpApi.User.AutoLogin() != nil {
		err := httpApi.User.Login(false)
		if err != nil {
			if err == user.ErrorPwd {
				zlog.Error("账号或密码错误,请重新输入账号密码")
				err = httpApi.User.Login(true)
			}
			if err != nil {
				zlog.Error(err.Error())
				return
			}
		}
	}
	c := command.Command{
		HttpApi: httpApi,
		Folder:  folder.NewFolder(),
	}
	// 初始化根目录
	files, err := c.HttpApi.GetFolder("0")
	if err != nil {
		zlog.Error(err.Error())
		return
	}
	folder.AddFolder(c.Folder, files)
	app := &cli.App{
		Name:    "Go-MiCloud",
		Usage:   "MiCloud Third Party Console Client Written By Golang",
		Version: "1.1",
		Commands: []*cli.Command{
			c.Login(),
			c.List(),
			c.Download(),
			c.Cd(),
			c.Upload(),
			c.Share(),
			c.Delete(),
			c.MkDir(),
			c.Tree(),
		},
		CommandNotFound: func(c *cli.Context, command string) {
			zlog.Error("命令不存在")
		},
	}
	line.CsLiner.SetWorldCompleter(nil)
	for {
		commandLine, err := line.CsLiner.Prompt()
		if err != nil {
			if err == liner.ErrPromptAborted || err == io.EOF {
				_ = line.CsLiner.Close()
				println("exit")
				return
			}
			zlog.Error(fmt.Sprintf("命令键入错误： %s", err.Error()))
			continue
		}
		line.CsLiner.AppendHistory(commandLine)
		var args = append(
			[]string{os.Args[0]},
			strings.Split(commandLine, " ")...)
		err = app.Run(args)
		if err != nil {
			zlog.Error(err.Error())
			continue
		}
	}
}
