package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/pkg/zlog"
	"os"
	"strings"
)

func (r *Command) Upload() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "Upload file",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			if args.Len() == 0 {
				return errors.New("缺少参数")
			}
			err := r.upload(args.Get(0))
			if err != nil {
				zlog.Error("上传失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) upload(name string) error {
	fileName := strings.ReplaceAll(name, "\\s", " ")
	fileInfo, err := os.Stat(fileName)
	if os.IsPermission(err) {
		return errors.New("没有访问权限")
	}
	if os.IsNotExist(err) {
		return errors.New("文件不存在")
	}
	if fileInfo.IsDir() {
		return errors.New("目前不支持上传文件夹")
	}
	zlog.Info("开始上传")
	_, err = r.HttpApi.UploadFile(fileName, r.Folder.Cursor.Id)
	if err != nil {
		return errors.New(fmt.Sprintf("上传失败！%s\n", err))
	} else {
		zlog.Info(fmt.Sprintf("[ %s ]上传成功", fileName))
	}
	return nil
}
