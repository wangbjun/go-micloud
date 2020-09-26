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

func (r *Command) DownloadAlbum() *cli.Command {
	return &cli.Command{
		Name:  "downloadAlbum",
		Usage: "下载相册照片文件",
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
			ablums, err := r.Request.GetAblums()
			if err != nil {
				return err
			}
			for _, v := range ablums {
				// 如果参数为空，则下载所有相册
				if fileName != "" {
					if v.Name == fileName {
						err = r.downloadAlbum(v.AlbumId, savePath+"/"+v.Name, 0)
					}
				} else {
					err = r.downloadAlbum(v.AlbumId, savePath+"/"+v.Name, 0)
				}
				if err != nil {
					zlog.PrintError("下载失败：" + err.Error())
				}
			}
			return nil
		},
	}
}

func (r *Command) downloadAlbum(albumId, saveDir string, page int) error {
	albumFiles, isLastPage, err := r.Request.GetAblumPhotos(albumId, page)
	if err != nil {
		return fmt.Errorf("获取相册照片失败: %w", err)
	}
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err = os.Mkdir(saveDir, 0755)
		if err != nil {
			return fmt.Errorf("创建相册目录失败: %w", err)
		}
	}
	// 最后一页，下载结束
	if albumFiles == nil && isLastPage == true {
		return nil
	}
	for _, f := range albumFiles {
		go func() {
			r.TaskManager.AddDownloadTask(&f, saveDir, api.TypeDownloadAlbum)
		}()
		zlog.PrintInfo(fmt.Sprintf("添加下载任务: %s", saveDir+"/"+f.Name))
		time.Sleep(time.Millisecond * 10)
	}
	time.Sleep(time.Second * 5)
	err = r.downloadAlbum(albumId, saveDir, page+1)
	if err != nil {
		return fmt.Errorf("相册下载失败: %w", err)
	}
	return nil
}
