package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
	"go-micloud/configs"
	"go-micloud/internal/api"
	"go-micloud/internal/command"
	"go-micloud/internal/user"
	"go-micloud/pkg/line"
	"go-micloud/pkg/zlog"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("用户目录不存在: %s\n", err.Error())
	}
	err = configs.Init(userHomeDir + "/.micloud.json")
	if err != nil {
		log.Panic(err.Error())
	}
	// 日志初始化
	zlog.Init(configs.Conf.LogFile)
	// 用户登录
	httpApi := api.New(user.New())
	if err := httpApi.User.Login(false); err != nil {
		zlog.PrintError("登录失败： " + err.Error())
		return
	}
	zlog.PrintInfo("登录成功")
	cmd := command.Command{
		Request:     httpApi,
		Folder:      api.NewFolder(),
		TaskManager: api.NewManager(httpApi),
		Liner:       line.NewLiner(),
	}
	if err := cmd.InitRoot(); err != nil {
		zlog.PrintError("初始化根目录失败： " + err.Error())
		return
	}
	app := &cli.App{
		Name:    "Go-MiCloud",
		Usage:   "MiCloud Third Party Console Client Written In Golang",
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
			cmd.Jobs(),
			cmd.Quit(),
			cmd.ListAlbum(),
			cmd.DownloadAlbum(),
		},
		CommandNotFound: func(c *cli.Context, command string) {
			zlog.PrintError(fmt.Sprintf("命令[ %s ]不存在", command))
		},
	}
	for {
		commandLine, err := cmd.Liner.Prompt()
		if err != nil {
			if err == liner.ErrPromptAborted || err == io.EOF {
				_ = cmd.Liner.Close()
				println("exit")
				return
			}
			zlog.PrintError(fmt.Sprintf("命令键入错误： %s", err.Error()))
			continue
		}
		var input = commandLine
		var argument = ""
		if strings.Contains(commandLine, " ") {
			i := strings.Index(commandLine, " ")
			input = commandLine[0:i]
			argument = commandLine[i+1:]
		}
		err = app.Run([]string{app.Name, input, argument})
		if err != nil {
			zlog.PrintError(err.Error())
			continue
		}
		cmd.Liner.AppendHistory(commandLine)
	}
}
