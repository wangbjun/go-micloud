package command

import (
	"errors"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/user"
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
			if r.FileApi.User.AutoLogin() != nil {
				err := r.FileApi.User.Login(false)
				if err != nil {
					if err == user.ErrorPwd {
						zlog.PrintError("账号或密码错误,请重新输入账号密码")
						err := r.FileApi.User.Login(true)
						if err != nil {
							return err
						}
					} else {
						return err
					}
				}
				_ = r.List().Run(ctx)
			}
			return nil
		},
	}
}
