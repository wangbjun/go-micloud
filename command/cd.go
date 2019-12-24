package command

import (
	"errors"
	"github.com/urfave/cli/v2"
	"go-micloud/lib/line"
	"strings"
)

func Cd() *cli.Command {
	return &cli.Command{
		Name:  "cd",
		Usage: "Change Directory",
		Action: func(context *cli.Context) error {
			var args = context.Args()
			if args.Len() == 0 || args.First() == "/" {
				line.CsLiner.RemoveDir(-1)
				DirList = DirList[0:1]
				return nil
			}
			var dir = args.First()
			if strings.HasPrefix(dir, "..") {
				count := strings.Count(dir, "..")
				line.CsLiner.RemoveDir(count)
				for i := 0; i < count; i++ {
					DirList = DirList[0 : len(DirList)-1]
				}
				return nil
			}
			dir = strings.ReplaceAll(dir, "\\s", " ")
			file, ok := FileMap[dir]
			if !ok || file.Type != "folder" {
				return errors.New("目录不存在")
			}
			DirList = append(DirList, file.Id)
			line.CsLiner.AppendDir(file.Name)
			_ = List().Run(context)
			return nil
		},
	}
}
