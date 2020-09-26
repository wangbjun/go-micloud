package command

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/configs"
	"go-micloud/internal/api"
	"go-micloud/pkg/zlog"
	"os"
	"strings"
	"time"
)

func (r *Command) Download() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "下载文件或者文件夹",
		Action: func(ctx *cli.Context) error {
			param := ctx.Args().First()
			if param == "" {
				return errors.New("缺少参数")
			}
			var (
				fileName = param
				savePath = configs.Conf.WorkDir
			)
			i := strings.Index(param, "-d")
			if i > 0 {
				fileName = param[:i-1]
				savePath = param[i+3:]
			}
			var fileInfo *api.File
			for _, f := range r.Folder.Cursor.Child {
				if f.Name == fileName {
					fileInfo = f
				}
			}
			if fileInfo == nil {
				return fmt.Errorf("文件[ %s ]不存在", fileName)
			}
			var err error
			if fileInfo.Type == "folder" {
				err = r.download(fileInfo, savePath+"/"+fileName)
			} else {
				err = r.download(fileInfo, savePath)
			}
			if err != nil {
				zlog.PrintError("下载失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) download(fileInfo *api.File, saveDir string) error {
	if fileInfo.Type == "folder" {
		files, err := r.Request.GetFolder(fileInfo.Id)
		if err != nil {
			return errors.New("获取目录信息失败")
		}
		if _, err := os.Stat(saveDir); os.IsNotExist(err) {
			err = os.Mkdir(saveDir, 0755)
			if err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
		}
		for _, f := range files {
			var err error
			if f.Type == "folder" {
				err = r.download(f, saveDir+"/"+f.Name)
			} else {
				err = r.download(f, saveDir)
			}
			if err != nil {
				zlog.PrintError(fmt.Sprintf("[ %s ]下载失败： %s", f.Name, err))
			}
		}
	} else {
		go func() {
			r.TaskManager.AddDownloadTask(fileInfo, saveDir, api.TypeDownload)
		}()
		zlog.PrintInfo(fmt.Sprintf("添加下载任务: %s", saveDir+"/"+fileInfo.Name))
		time.Sleep(time.Millisecond * 10)
	}
	return nil
}
