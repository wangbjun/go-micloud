package command

import (
	"github.com/urfave/cli/v2"
)

func (r *Command) Jobs() *cli.Command {
	return &cli.Command{
		Name:  "jobs",
		Usage: "展示后台当前所有下载和上传任务",
		Action: func(ctx *cli.Context) error {
			r.TaskManager.ShowTask()
			return nil
		},
	}
}
