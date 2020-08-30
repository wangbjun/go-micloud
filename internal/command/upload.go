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
			for i := 0; i < args.Len(); i++ {
				fileName := args.Get(i)
				fileName = strings.ReplaceAll(fileName, "\\s", " ")
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
				_, err = r.HttpApi.UploadFile(fileName, DirList[len(DirList)-1])
				if err != nil {
					return errors.New(fmt.Sprintf("上传失败！%s\n", err))
				} else {
					zlog.Info(fmt.Sprintf("[ %s ]上传成功", fileName))
				}
			}
			return nil
		},
	}
}
