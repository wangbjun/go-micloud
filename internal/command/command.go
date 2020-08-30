package command

import (
	"go-micloud/internal/api"
	"go-micloud/internal/folder"
)

type Command struct {
	HttpApi api.Api
	Folder  *folder.Folder
}
