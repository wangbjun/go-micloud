package command

import (
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
	"go-micloud/configs"
	"go-micloud/internal/api"
	"go-micloud/pkg/color"
	"go-micloud/pkg/zlog"
	"io"
	"os"
	"strings"
)

func (r *Command) Download() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "Download file",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			if args.Len() == 0 {
				return errors.New("缺少参数")
			}
			var fileName = strings.ReplaceAll(args.Get(0), "\\s", " ")
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
				zlog.Error("下载失败：" + err.Error())
			}
			return nil
		},
	}
}

func (r *Command) download(fileInfo *api.File, dir string) error {
	if fileInfo.Type == "folder" {
		files, err := r.HttpApi.GetFolder(fileInfo.Id)
		if err != nil {
			return errors.New("获取目录信息失败")
		}
		if _, err := os.Stat(configs.WorkDir + "/" + dir); os.IsNotExist(err) {
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
				zlog.Error(fmt.Sprintf("[ %s ]下载失败： %s", f.Name, err))
			}
		}
	} else {
		zlog.Info(fmt.Sprintf("[ %s ]开始下载", fileInfo.Name))
		filePath := configs.WorkDir + "/" + dir + "/" + fileInfo.Name
		if fs, err := os.Stat(filePath); err == nil && fs.Size() == fileInfo.Size {
			return errors.New("文件已存在，跳过")
		}
		openFile, err := os.Create(filePath)
		if err != nil {
			return errors.New("创建失败: " + err.Error())
		}
		file, err := r.HttpApi.GetFile(fileInfo.Id)
		if err != nil {
			return err
		}
		_, err = io.Copy(openFile, io.TeeReader(file, &WriteCounter{FileSize: uint64(fileInfo.Size)}))
		fmt.Printf("\n")
		if err != nil {
			return errors.New("写入失败: " + err.Error())
		}
		zlog.Info(fmt.Sprintf("[ %s ]下载成功", fileInfo.Name))
	}
	return nil
}

type WriteCounter struct {
	Total    uint64
	FileSize uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\r"+color.Green("### Info: 下载中... %s/%s"), humanize.Bytes(wc.Total), humanize.Bytes(wc.FileSize))
}
