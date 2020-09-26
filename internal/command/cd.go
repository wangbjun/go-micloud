package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/api"
	"go-micloud/pkg/line"
	"go-micloud/pkg/zlog"
	"strings"
)

type Command struct {
	Request     *api.Api
	Folder      *api.Folder
	TaskManager *api.Manager
	Liner       *line.Liner
}

func (r *Command) Cd() *cli.Command {
	return &cli.Command{
		Name:  "cd",
		Usage: "改变当前目录，例如：cd movies",
		Action: func(ctx *cli.Context) error {
			dirName := ctx.Args().First()
			if strings.Trim(dirName, " ") == "/" || strings.Trim(dirName, " ") == "" {
				dirName = "/"
			}
			err := r.Folder.ChangeFolder(dirName)
			if err != nil {
				return err
			}
			files, err := r.Request.GetFolder(r.Folder.Cursor.Id)
			if err != nil {
				return err
			}
			r.Folder.AddFolder(files)
			r.setUpWordCompleter(files)
			r.setUpLinePrefix(r.Folder.Cursor)
			return nil
		},
	}
}

// 初始化根目录
func (r *Command) InitRoot() error {
	files, err := r.Request.GetFolder("0")
	if err != nil {
		zlog.PrintError(err.Error())
		return err
	}
	r.Folder.AddFolder(files)
	r.setUpWordCompleter(files)
	return nil
}

// 设置Tab补全提示
func (r *Command) setUpWordCompleter(files []*api.File) {
	var completerWord []string
	for _, f := range files {
		completerWord = append(completerWord, f.Name)
	}
	r.Liner.SetWorldCompleter(completerWord)
}

// 设置命令行前缀
func (r *Command) setUpLinePrefix(cursor *api.File) {
	var (
		names []string
		c     = cursor
	)
	for c != nil {
		if c.Name != "/" {
			names = append(names, c.Name)
		}
		c = c.Parent
	}
	var path string
	for i := len(names); i > 0; i-- {
		path = path + "/" + names[i-1]
	}
	r.Liner.SetUpPrefix(strings.ReplaceAll(path, "//", "/"))
}
