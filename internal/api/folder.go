package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/url"
)

const (
	GetFolders   = BaseUri + "/drive/user/folders/%s/children"
	CreateFolder = BaseUri + "/drive/v2/user/folders/create"
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

func (api *Api) CreateFolder(name, parentId string) (*File, error) {
	resp, err := api.User.HttpClient.PostForm(CreateFolder, url.Values{
		"name":         []string{name},
		"parentId":     []string{parentId},
		"serviceToken": []string{api.User.ServiceToken},
	})
	if err != nil {
		return nil, err
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	msg := &MsgV2{}
	err = json.Unmarshal(result, msg)
	if err != nil {
		return nil, err
	}
	if msg.Result == "ok" {
		return &msg.Data, nil
	} else {
		if gjson.Get(string(result), "R").Int() == 401 {
			return nil, errors.New("登录授权失败")
		}
		log.Printf("%s\n", result)
		return nil, errors.New("创建目录失败")
	}
}

type DeleteFile struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func (api *Api) DeleteFile(id, fType string) error {
	record := []DeleteFile{{
		Id:   id,
		Type: fType,
	}}
	content, _ := json.Marshal(record)
	resp, err := api.User.HttpClient.PostForm(DeleteFiles, url.Values{
		"operateType":    []string{"DELETE"},
		"operateRecords": []string{string(content)},
		"serviceToken":   []string{api.User.ServiceToken},
	})
	if err != nil {
		return err
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	msg := &MsgV2{}
	err = json.Unmarshal(result, msg)
	if err != nil {
		return err
	}
	if msg.Result == "ok" {
		return nil
	} else {
		if gjson.Get(string(result), "R").Int() == 401 {
			return errors.New("登录授权失败")
		}
		log.Printf("%s\n", result)
		return errors.New("创建目录失败")
	}
}
