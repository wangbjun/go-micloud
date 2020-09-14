package command

import (
	"github.com/urfave/cli/v2"
	"go-micloud/internal/folder"
)

func (r *Command) Tree() *cli.Command {
	return &cli.Command{
		Name:  "tree",
		Usage: "打印树型目录结构",
		Action: func(ctx *cli.Context) error {
			folder.PrintFolder(r.Folder.Cursor, 0)
			return nil
		},
	}
}
