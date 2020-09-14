package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/file"
	"go-micloud/pkg/zlog"
	"strings"
)

func (r *Command) Delete() *cli.Command {
	return &cli.Command{
		Name:  "rm",
		Usage: "删除文件或者文件夹，即放入回收站",
		Action: func(ctx *cli.Context) error {
			var fileName = strings.Join(ctx.Args().Slice(), " ")
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
			err := r.HttpApi.DeleteFile(fileInfo.Id, fileInfo.Type)
			if err != nil {
				zlog.Error("删除失败：" + err.Error())
				return err
			}
			zlog.Info(fmt.Sprintf("[ %s ]删除成功", fileName))
			return nil
		},
	}
}
