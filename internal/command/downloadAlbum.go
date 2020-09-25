package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/configs"
	"go-micloud/internal/api"
	"go-micloud/pkg/zlog"
	"os"
	"time"
)

func (r *Command) DownloadAlbum() *cli.Command {
	return &cli.Command{
		Name:  "downloadAlbum",
		Usage: "下载相册照片文件",
		Action: func(ctx *cli.Context) error {
			albumName := ctx.Args().First()
			ablums, err := r.FileApi.GetAblums()
			if err != nil {
				return err
			}
			for _, v := range ablums {
				// 如果参数为空，则下载所有相册
				if albumName != "" {
					if v.Name == albumName {
						err = r.downloadAlbum(v.AlbumId, v.Name, 0)
					}
				} else {
					err = r.downloadAlbum(v.AlbumId, v.Name, 0)
				}
				if err != nil {
					zlog.PrintError("下载失败：" + err.Error())
				}
			}
			return nil
		},
	}
}

func (r *Command) downloadAlbum(albumId, albumName string, page int) error {
	albumFiles, isLastPage, err := r.FileApi.GetAblumPhotos(albumId, page)
	if err != nil {
		return fmt.Errorf("获取相册照片失败: %w", err)
	}
	if _, err := os.Stat(configs.Conf.WorkDir + "/" + albumName); os.IsNotExist(err) {
		err = os.Mkdir(albumName, 0755)
		if err != nil {
			return fmt.Errorf("创建相册目录失败: %w", err)
		}
	}
	for _, f := range albumFiles {
		go func() {
			r.TaskManage.AddDownloadTask(&f, albumName, api.TypeDownloadAlbum)
		}()
		zlog.PrintInfo(fmt.Sprintf("添加下载任务: %s", albumName+"/"+f.Name))
		time.Sleep(time.Millisecond * 10)
	}
	if isLastPage != true {
		err := r.downloadAlbum(albumId, albumName, page+1)
		if err != nil {
			return fmt.Errorf("相册下载失败: %w", err)
		}
		time.Sleep(time.Second)
	}
	return nil
}
