package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/folder"
)

func (r *Command) List() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "List all files",
		Action: func(context *cli.Context) error {
			files, err := r.HttpApi.GetFolder(r.Folder.Cursor.Id)
			if err != nil {
				return err
			}
			folder.AddFolder(r.Folder, r.Folder.Cursor.Name, files)
			folder.Format(r.Folder.Cursor.Child)
			return nil
		},
	}
}
