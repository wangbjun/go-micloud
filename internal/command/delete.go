package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/file"
	"go-micloud/pkg/zlog"
)

func (r *Command) Delete() *cli.Command {
	return &cli.Command{
		Name:  "rm",
		Usage: "删除文件或者文件夹，即放入回收站",
		Action: func(ctx *cli.Context) error {
			fileName := ctx.Args().First()
			if fileName == "" {
				return errors.New("缺少参数")
			}
			var fileInfo *file.File
			for _, f := range r.Folder.Cursor.Child {
				if f.Name == fileName {
					fileInfo = f
				}
			}
			if fileInfo == nil {
				return errors.New("当前目录不存在该文件")
			}
			err := r.FileApi.DeleteFile(fileInfo.Id, fileInfo.Type)
			if err != nil {
				return err
			}
			zlog.PrintInfo(fmt.Sprintf("[ %s ]删除成功", fileName))
			return nil
		},
	}
}
