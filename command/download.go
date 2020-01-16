package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"go-micloud/config"
	"os"
	"strings"
)

func Download() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "Download file",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			for i := 0; i < args.Len(); i++ {
				fileName := args.Get(i)
				fileName = strings.ReplaceAll(fileName, "\\s", " ")
				fileInfo, ok := FileMap[fileName]
				if !ok {
					fmt.Println("===> 当前目录不存在该文件！")
					continue
				}
				if fileInfo.Type == "folder" {
					fmt.Println("===> 目前不支持下载文件夹！")
					continue
				}
				fmt.Println("===> 开始下载！")
				file, err := api.FileApi.GetFile(fileInfo.Id)
				if err != nil {
					fmt.Printf("===> 下载失败！Error: %s\n", err)
					continue
				}
				filePath := config.WorkDir + "/" + fileName
				openFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					fmt.Println("===> 创建失败！")
					continue
				}
				_, err = openFile.Write(file)
				if err != nil {
					fmt.Println("===> 写入失败！")
					continue
				}
				fmt.Println("===> 下载成功！")
			}
			return nil
		},
	}
}
