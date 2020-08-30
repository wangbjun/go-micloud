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
			for i := 0; i < args.Len(); i++ {
				var fileName = strings.ReplaceAll(args.Get(i), "\\s", " ")
				var fileInfo *api.File
				for _, f := range r.Folder.Cursor.Child {
					if f.Name == fileName {
						fileInfo = f
					}
				}
				if fileInfo == nil {
					return errors.New("当前目录不存在该文件")
				}
				if fileInfo.Type == "folder" {
					return errors.New("目前不支持下载文件夹")
				}
				zlog.Info("开始下载")
				file, err := r.HttpApi.GetFile(fileInfo.Id)
				if err != nil {
					return err
				}
				filePath := configs.WorkDir + "/" + fileName
				openFile, err := os.Create(filePath)
				if err != nil {
					return errors.New("创建失败: " + err.Error())
				}
				_, err = io.Copy(openFile, io.TeeReader(file, &WriteCounter{FileSize: uint64(fileInfo.Size)}))
				fmt.Printf("\n")
				if err != nil {
					return errors.New("写入失败: " + err.Error())
				}
				zlog.Info(fmt.Sprintf("[ %s ]下载成功", fileName))
			}
			return nil
		},
	}
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
