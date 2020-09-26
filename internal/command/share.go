package command

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/api"
	"go-micloud/pkg/zlog"
	"net/url"
	"strings"
)

func (r *Command) Share() *cli.Command {
	return &cli.Command{
		Name:  "share",
		Usage: "获取一个公共分享链接",
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
			if fileInfo.Type == "folder" {
				return errors.New("目前不支持分享文件夹")
			}
			downloadUrl, err := r.Request.GetFileDownLoadUrl(fileInfo.Id)
			if err != nil {
				return fmt.Errorf("获取下载地址失败：%s", err)
			}
			resp, err := r.Request.PostForm("http://t.wibliss.com/api/v1/create", url.Values{"url": []string{downloadUrl}})
			if err != nil {
				return fmt.Errorf("获取短网址失败：%w", err)
			}
			shortUrl := gjson.Get(string(resp), "data.url").String()
			if shortUrl == "" {
				return fmt.Errorf("获取短网址失败：%s", resp)
			}
			i := strings.LastIndex(shortUrl, "/")
			shortUrl = shortUrl[:i] + "?t=" + shortUrl[i+1:]
			zlog.PrintInfo(fmt.Sprintf("获取分享成功,有效期24小时，复制链接( %s )到浏览器里面打开下载,请注意浏览器弹框！", shortUrl))
			return nil
		},
	}
}
