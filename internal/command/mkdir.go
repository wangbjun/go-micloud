package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/pkg/zlog"
)

func (r *Command) MkDir() *cli.Command {
	return &cli.Command{
		Name:  "mkdir",
		Usage: "创建目录",
		Action: func(ctx *cli.Context) error {
			fileName := ctx.Args().First()
			if fileName == "" {
				return errors.New("缺少参数")
			}
			_, err := r.FileApi.CreateFolder(fileName, r.Folder.Cursor.Id)
			if err != nil {
				zlog.PrintError("创建目录失败：" + err.Error())
				return err
			}
			zlog.PrintInfo(fmt.Sprintf("[ %s ]创建成功", fileName))
			return nil
		},
	}
}
