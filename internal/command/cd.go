package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/folder"
	"strings"
)

func (r *Command) Cd() *cli.Command {
	return &cli.Command{
		Name:  "cd",
		Usage: "Change Directory",
		Action: func(context *cli.Context) error {
			var (
				dir = context.Args().First()
				err error
			)
			if strings.Trim(dir, " ") == "/" || strings.Trim(dir, " ") == "" {
				dir = "/"
			}
			err = folder.ChangeFolder(r.Folder, dir)
			if err != nil {
				return err
			}
			files, err := r.HttpApi.GetFolder(r.Folder.Cursor.Id)
			if err != nil {
				return err
			}
			folder.AddFolder(r.Folder, r.Folder.Cursor.Name, files)

			return nil
		},
	}
}
