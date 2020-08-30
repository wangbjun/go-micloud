package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/api"
	"go-micloud/internal/folder"
)

var DirList []string

var FileMap map[string]*api.File

func (r *Command) List() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "List all files",
		Action: func(context *cli.Context) error {
			folder.Format(r.Folder.Cursor.Child)
			return nil
		},
	}
}
