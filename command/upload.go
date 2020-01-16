package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"os"
	"strings"
)

func Upload() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "Upload file",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			for i := 0; i < args.Len(); i++ {
				fileName := args.Get(i)
				fileName = strings.ReplaceAll(fileName, "\\s", " ")
				fileInfo, err := os.Stat(fileName)
				if os.IsPermission(err) {
					fmt.Println("===> 没有访问权限！")
					continue
				}
				if os.IsNotExist(err) {
					fmt.Println("===> 文件不存在！")
					continue
				}
				if fileInfo.IsDir() {
					fmt.Println("===> 目前不支持上传文件夹！")
					continue
				}
				fmt.Println("===> 开始上传！")
				_, err = api.FileApi.UploadFile(fileName, DirList[len(DirList)-1])
				if err != nil {
					fmt.Printf("===> 上传失败！Error: %s\n", err)
				} else {
					fmt.Println("===> 上传成功！")
				}
			}
			return nil
		},
	}
}
