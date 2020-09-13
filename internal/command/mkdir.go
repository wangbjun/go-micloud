package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/pkg/zlog"
)

func (r *Command) MkDir() *cli.Command {
	return &cli.Command{
		Name:            "mkdir",
		Usage:           "创建目录",
		SkipFlagParsing: true,
		Action: func(context *cli.Context) error {
			var fileName = context.Args().First()
			if fileName == "" {
				return errors.New("缺少参数")
			}
			_, err := r.HttpApi.CreateFolder(fileName, r.Folder.Cursor.Id)
			if err != nil {
				zlog.Error("创建目录失败：" + err.Error())
				return err
			}
			zlog.Info(fmt.Sprintf("[ %s ]创建成功", fileName))
			return nil
		},
	}
}
