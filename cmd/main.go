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
	"strings"
)

func main() {
	httpApi := file.NewApi(user.NewUser())
	if !userLogin(httpApi) {
		return
	}
	cmd := command.Command{
		FileApi:    httpApi,
		Folder:     folder.NewFolder(),
		TaskManage: file.NewManage(httpApi),
	}
	if !initFolder(cmd) {
		return
	}
	app := &cli.App{
		Name:    "Go-MiCloud",
		Usage:   "MiCloud Third Party Console Client Written By Golang",
		Version: "1.2",
		Commands: []*cli.Command{
			cmd.Login(),
			cmd.List(),
			cmd.Download(),
			cmd.Cd(),
			cmd.Upload(),
			cmd.Share(),
			cmd.Delete(),
			cmd.MkDir(),
			cmd.Tree(),
		},
		CommandNotFound: func(c *cli.Context, command string) {
			zlog.PrintError(fmt.Sprintf("命令[ %s ]不存在", command))
		},
	}
	for {
		commandLine, err := line.CsLiner.Prompt()
		if err != nil {
			if err == liner.ErrPromptAborted || err == io.EOF {
				_ = line.CsLiner.Close()
				println("exit")
				return
			}
			zlog.PrintError(fmt.Sprintf("命令键入错误： %s", err.Error()))
			continue
		}
		var cmd = commandLine
		var argument = ""
		if strings.Contains(commandLine, " ") {
			i := strings.Index(commandLine, " ")
			cmd = commandLine[0:i]
			argument = commandLine[i+1:]
		}
		err = app.Run([]string{app.Name, cmd, argument})
		if err != nil {
			zlog.PrintError(err.Error())
			continue
		}
		line.CsLiner.AppendHistory(commandLine)
	}
}

// 初始化根目录
func initFolder(c command.Command) bool {
	files, err := c.FileApi.GetFolder("0")
	if err != nil {
		zlog.PrintError(err.Error())
		return false
	}
	folder.AddFolder(c.Folder, files)
	return true
}

// 用户登录
func userLogin(httpApi *file.Api) bool {
	if httpApi.User.AutoLogin() == nil {
		zlog.PrintInfo("自动登录成功！")
		return true
	}
	err := httpApi.User.Login(false)
	if err != nil {
		if err == user.ErrorPwd {
			zlog.PrintError("账号或密码错误,请重试！")
			err = httpApi.User.Login(true)
		}
		if err != nil {
			zlog.PrintError("登录失败：" + err.Error())
			return false
		}
	}
	zlog.PrintInfo("登录成功！")
	return true
}
