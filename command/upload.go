package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"os"
)

func Upload() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "Upload file",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			for i := 0; i < args.Len(); i++ {
				fileName := args.Get(i)
				fileInfo, err := os.Stat(fileName)
				if os.IsPermission(err) {
					fmt.Printf("===> 没有访问权限：%s\n", fileName)
					continue
				}
				if os.IsNotExist(err) {
					fmt.Printf("===> 文件不存在：%s\n", fileName)
					continue
				}
				if fileInfo.IsDir() {
					fmt.Printf("===> 目前不支持上传文件夹：%s\n", fileName)
					continue
				}
				fmt.Printf("===> [ %s ]开始上传！\n", fileName)
				_, err = api.FileApi.UploadFile(fileName, DirList[len(DirList)-1])
				if err != nil {
					panic(err)
					fmt.Printf("===> [ %s ]上传失败！\n", fileName)
				} else {
					fmt.Printf("===> [ %s ]上传成功！\n", fileName)
				}
			}
			return nil
		},
	}
}
