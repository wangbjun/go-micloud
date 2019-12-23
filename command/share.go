package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
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
					fmt.Printf("===> 当前目录不存在该文件：%s\n", fileName)
					continue
				}
				if fileInfo.Type == "folder" {
					fmt.Printf("===> 目前不支持分享文件夹：%s\n", fileName)
					continue
				}
				downloadUrl, err := api.FileApi.GetFileDownLoadUrl(fileInfo.Id)
				if err != nil {
					fmt.Printf("===> [ %s ]获取失败！\n", fileName)
					continue
				}
				fmt.Println("===> 获取链接成功(有效期24小时): ")
				fmt.Println(downloadUrl)
			}
			return nil
		},
	}
}
