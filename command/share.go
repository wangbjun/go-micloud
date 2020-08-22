package command

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"go-micloud/lib/zlog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func Share() *cli.Command {
	return &cli.Command{
		Name:  "share",
		Usage: "Get public share url",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			for i := 0; i < args.Len(); i++ {
				fileName := args.Get(i)
				fileInfo, ok := FileMap[fileName]
				if !ok {
					fmt.Printf("===> 当前目录不存在该文件！\n")
					continue
				}
				if fileInfo.Type == "folder" {
					fmt.Printf("===> 目前不支持分享文件夹！\n")
					continue
				}
				downloadUrl, err := api.FileApi.GetFileDownLoadUrl(fileInfo.Id)
				if err != nil {
					fmt.Printf("===> 获取失败！Error: %s\n", err)
					continue
				}
				var shortUrl = downloadUrl
				resp, err := http.PostForm("http://t.wibliss.com/api/v1/create", url.Values{"url": []string{downloadUrl}})
				if err == nil {
					all, _ := ioutil.ReadAll(resp.Body)
					dataUrl := gjson.Get(string(all), "data.url").String()
					if dataUrl != "" {
						shortUrl = dataUrl
					}
					resp.Body.Close()
				}
				zlog.Logger.Info(shortUrl)

				i := strings.LastIndex(shortUrl, "/")

				shortUrl = shortUrl[:i] + "?t=" + shortUrl[i+1:]

				fmt.Println("===> 获取分享链接成功,点击链接跳转到下载页面，请注意浏览器弹框！(有效期24小时): " + shortUrl)
			}
			return nil
		},
	}
}
