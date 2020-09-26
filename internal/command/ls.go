package command

import (
	"github.com/urfave/cli/v2"
)

func (r *Command) List() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "列表当前目录所有文件和文件夹",
		Action: func(ctx *cli.Context) error {
			files, err := r.Request.GetFolder(r.Folder.Cursor.Id)
			if err != nil {
				return err
			}
			r.Folder.AddFolder(files)
			r.Folder.Format()
			r.setUpWordCompleter(files)
			return nil
		},
	}
}
