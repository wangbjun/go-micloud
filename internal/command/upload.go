package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/file"
	"go-micloud/pkg/zlog"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

func (r *Command) Upload() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "上传文件或者文件夹",
		Action: func(context *cli.Context) error {
			var fileName = context.Args().First()
			if fileName == "" {
				return errors.New("缺少参数")
			}
			err := r.upload(fileName, r.Folder.Cursor.Id)
			if err != nil {
				zlog.Error("上传失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) upload(fileName, parentId string) error {
	fileInfo, err := os.Stat(fileName)
	if os.IsPermission(err) {
		return errors.New("没有访问权限")
	}
	if os.IsNotExist(err) {
		return errors.New("文件不存在")
	}
	if fileInfo.IsDir() {
		folderId, err := r.HttpApi.CreateFolder(path.Base(fileName), parentId)
		if err != nil {
			return err
		}
		dir, err := ioutil.ReadDir(fileName)
		for _, d := range dir {
			if strings.HasPrefix(d.Name(), ".") {
				continue
			}
			err := r.upload(fileName+"/"+d.Name(), folderId)
			if err != nil {
				return err
			}
		}
	} else {
		zlog.Info(fmt.Sprintf("[ %s ]开始上传", fileName))
		time.Sleep(time.Millisecond * 100)
		_, err = r.HttpApi.UploadFile(fileName, parentId)
		if err != nil {
			if err == file.SizeTooBigError {
				zlog.Error(fmt.Sprintf("[ %s ] %s", fileName, err))
				return nil
			}
			return err
		} else {
			zlog.Info(fmt.Sprintf("[ %s ]上传成功", fileName))
		}
	}
	return nil
}
