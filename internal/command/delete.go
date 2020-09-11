package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/api"
	"go-micloud/pkg/zlog"
)

func (r *Command) Delete() *cli.Command {
	return &cli.Command{
		Name:            "rm",
		Usage:           "delete file",
		SkipFlagParsing: true,
		Action: func(context *cli.Context) error {
			var fileName = context.Args().First()
			if fileName == "" {
				return errors.New("缺少参数")
			}
			var fileInfo *api.File
			for _, f := range r.Folder.Cursor.Child {
				if f.Name == fileName {
					fileInfo = f
				}
			}
			if fileInfo == nil {
				return errors.New("当前目录不存在该文件")
			}
			err := r.HttpApi.DeleteFile(fileInfo.Id, fileInfo.Type)
			if err != nil {
				zlog.Error("删除失败：" + err.Error())
			}
			zlog.Info(fmt.Sprintf("[ %s ]删除成功", fileName))
			return nil
		},
	}
}
