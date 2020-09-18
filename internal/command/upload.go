package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
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
		Action: func(ctx *cli.Context) error {
			filePath := ctx.Args().First()
			if filePath == "" {
				return errors.New("缺少参数")
			}
			err := r.upload(filePath, r.Folder.Cursor.Id)
			if err != nil {
				zlog.PrintError("上传失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) upload(filePath, parentId string) error {
	fileInfo, err := os.Stat(filePath)
	if os.IsPermission(err) {
		return errors.New("没有访问权限")
	}
	if os.IsNotExist(err) {
		return errors.New("文件不存在")
	}
	if fileInfo.IsDir() {
		folderId, err := r.FileApi.CreateFolder(path.Base(filePath), parentId)
		if err != nil {
			return err
		}
		dir, err := ioutil.ReadDir(filePath)
		for _, d := range dir {
			if strings.HasPrefix(d.Name(), ".") {
				continue
			}
			err := r.upload(filePath+"/"+d.Name(), folderId)
			if err != nil {
				return err
			}
		}
	} else {
		go func() {
			r.TaskManage.AddUploadTask(filePath, parentId)
		}()
		zlog.PrintInfo(fmt.Sprintf("添加上传任务: %s", filePath))
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}
