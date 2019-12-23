package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/user"
)

func Login() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Login account",
		Action: func(context *cli.Context) error {
			if user.Account.LoginManual() == nil {
				fmt.Println("===> 自动登录成功！")
				_ = List().Run(context)
			} else {
				err := user.Account.Login(false)
				if err != nil {
					if err == user.ErrorPwd {
						fmt.Println("===> 账号或密码错误,请重新输入账号密码！")
						err := user.Account.Login(true)
						if err != nil {
							return err
						}
					} else {
						return err
					}
				}
				_ = List().Run(context)
			}
			return nil
		},
	}
}
