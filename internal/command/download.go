package command

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/configs"
	"go-micloud/internal/api"
	"go-micloud/pkg/zlog"
	"os"
	"time"
)

func (r *Command) Download() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "下载文件或者文件夹",
		Action: func(ctx *cli.Context) error {
			fileName := ctx.Args().First()
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
			var err error
			if fileInfo.Type == "folder" {
				err = r.download(fileInfo, fileName)
			} else {
				err = r.download(fileInfo, "")
			}
			if err != nil {
				zlog.PrintError("下载失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) download(fileInfo *api.File, dir string) error {
	if fileInfo.Type == "folder" {
		files, err := r.FileApi.GetFolder(fileInfo.Id)
		if err != nil {
			return errors.New("获取目录信息失败")
		}
		if _, err := os.Stat(configs.Conf.WorkDir + "/" + dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				return errors.New("创建目录失败")
			}
		}
		for _, f := range files {
			var err error
			if f.Type == "folder" {
				err = r.download(f, dir+"/"+f.Name)
			} else {
				err = r.download(f, dir)
			}
			if err != nil {
				zlog.PrintError(fmt.Sprintf("[ %s ]下载失败： %s", f.Name, err))
			}
		}
	} else {
		go func() {
			r.TaskManage.AddDownloadTask(fileInfo, dir, api.TypeDownload)
		}()
		zlog.PrintInfo(fmt.Sprintf("添加下载任务: %s", dir+"/"+fileInfo.Name))
		time.Sleep(time.Millisecond * 10)
	}
	return nil
}
