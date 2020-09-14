package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/folder"
)

func (r *Command) List() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "列表当前目录所有文件和文件夹",
		Action: func(ctx *cli.Context) error {
			files, err := r.HttpApi.GetFolder(r.Folder.Cursor.Id)
			if err != nil {
				return err
			}
			folder.AddFolder(r.Folder, files)
			folder.Format(r.Folder.Cursor.Child)
			return nil
		},
	}
}
