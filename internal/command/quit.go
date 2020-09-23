package command

import (
	"github.com/urfave/cli/v2"
	"os"
)

func (r *Command) Quit() *cli.Command {
	return &cli.Command{
		Name:    "quit",
		Aliases: []string{"exit"},
		Usage:   "退出应用",
		Action: func(ctx *cli.Context) error {
			_ = r.Liner.Close()
			os.Exit(0)
			return nil
		},
	}
}
