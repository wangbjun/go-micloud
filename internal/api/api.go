package api

import "go-micloud/internal/user"

type Api struct {
	User *user.User
}

func NewApi(user *user.User) *Api {
	api := Api{
		User: user,
	}
	return &api
}
