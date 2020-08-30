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

// 获取目录下的文件
func (api *Api) GetFolder(id string) ([]*File, error) {
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
			return nil, errors.New("登录授权失败")
		}
		return nil, errors.New("获取文件夹下文件失败")
	}
}
