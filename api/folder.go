package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

const (
	GetFolders   = BaseUri + "/drive/user/folders/%s/children"
	CreateFolder = BaseUri + "/drive/user/folders"
	DeleteFolder = BaseUri + "/drive/user/folders/%s/delete"
)

var ErrorNotLogin = errors.New("请登录,命令: login")

// 获取目录下的文件
func (api *api) GetFolder(id string) ([]*File, error) {
	apiUrl := fmt.Sprintf(GetFolders, id)
	result, err := api.get(apiUrl)
	if err != nil {
		return nil, err
	}
	msg := &Msg{}
	err = json.Unmarshal(result, msg)
	if err != nil {
		return nil, err
	}
	if msg.Result == "ok" {
		return msg.Data.List, nil
	} else {
		if gjson.Get(string(result), "R").Int() == 401 {
			return nil, ErrorNotLogin
		}
		return nil, errors.New("get folders failed")
	}
}
