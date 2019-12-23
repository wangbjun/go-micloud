package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"go-micloud/config"
	"os"
)

func Download() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "Download file",
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
					fmt.Printf("===> 目前不支持下载文件夹：%s\n", fileName)
					continue
				}
				fmt.Printf("===> [ %s ]开始下载！\n", fileName)
				file, err := api.FileApi.GetFile(fileInfo.Id)
				if err != nil {
					fmt.Printf("===> [ %s ]下载失败！Error: %s\n", fileName, err)
					continue
				}
				filePath := config.WorkDir + "/" + fileName
				openFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					fmt.Printf("===> [ %s ]创建失败!\n", filePath)
					continue
				}
				_, err = openFile.Write(file)
				if err != nil {
					fmt.Printf("===> [ %s ]写入失败\n", filePath)
					continue
				}
				fmt.Printf("===> [ %s ]下载成功！\n", fileName)
			}
			return nil
		},
	}
}
