package command

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
	"go-micloud/api"
	"go-micloud/lib/color"
	"go-micloud/lib/function"
	"go-micloud/lib/line"
)

var DirList []string

var FileMap map[string]*api.File

func List() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "List all files",
		Action: func(context *cli.Context) error {
			var folderId = "0"
			dirNum := len(DirList)
			if dirNum > 0 {
				folderId = DirList[dirNum-1]
			}
			files, err := api.FileApi.GetFolder(folderId)
			if err != nil {
				return err
			}
			format(files)

			if dirNum == 0 || DirList[dirNum-1] != folderId {
				DirList = append(DirList, folderId)
			}
			return nil
		},
	}
}

func format(files []*api.File) {
	fmt.Printf("total %d\n", len(files))
	FileMap = make(map[string]*api.File, 0)
	var words []string
	for _, v := range files {
		if v.Type == "file" {
			fmt.Printf("- | %-6s | %s | %s\n", humanize.Bytes(uint64(v.Size)), function.FormatTimeInt(int64(v.CreateTime), true), v.Name)
		} else {
			fmt.Printf("d | ------ | %s | %s\n", function.FormatTimeInt(int64(v.CreateTime), true), color.Blue(v.Name))
		}
		FileMap[v.Name] = v
		words = append(words, v.Name)
	}
	line.CsLiner.SetWorldCompleter(words)
}
