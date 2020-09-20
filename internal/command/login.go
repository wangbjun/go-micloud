package command

import (
	"errors"
	"github.com/urfave/cli/v2"
	"go-micloud/pkg/zlog"
)

func (r *Command) Login() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "登录小米云服务账号",
		Action: func(ctx *cli.Context) error {
			if r.FileApi.User.IsLogin {
				return errors.New("您已登录，账号为：" + r.FileApi.User.UserName)
			}
			err := r.FileApi.User.Login()
			if err != nil {
				zlog.PrintError("登录失败: " + err.Error())
				return nil
			}
			_ = r.List().Run(ctx)
			return nil
		},
	}
}
