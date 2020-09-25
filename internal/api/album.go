package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/url"
	"strings"
	"time"
)

type Album struct {
	AlbumId        string `json:"albumId"`
	LastUpdateTime int64  `json:"lastUpdateTime"`
	MediaCount     int    `json:"mediaCount"`
	Name           string `json:"name"`
}

// 获取相册列表
func (api *Api) GetAblums() ([]*Album, error) {
	result, err := api.get(fmt.Sprintf(GetAlbums, time.Now().UnixNano()))
	if err != nil {
		return nil, err
	}
	var ablums = make([]*Album, 0)
	err = json.Unmarshal([]byte(gjson.Get(string(result), "data.albums").Raw), &ablums)
	if err != nil {
		return ablums, err
	}
	for _, v := range ablums {
		switch v.AlbumId {
		case "1":
			v.Name = "相册"
		case "2":
			v.Name = "截屏"
		case "1000":
			v.Name = "私密"
		}
	}
	return ablums, nil
}

// 获取相册列表
func (api *Api) GetAblumPhotos(albumId string, page int) ([]File, bool, error) {
	result, err := api.get(fmt.Sprintf(GetPhotos, time.Now().UnixNano(), page, albumId))
	if err != nil {
		return nil, false, err
	}
	var (
		isLastPage = gjson.Get(string(result), "data.isLastPage").Bool()
		galleries  = gjson.Get(string(result), "data.galleries").Array()
	)
	if len(galleries) == 0 {
		return nil, isLastPage, nil
	}
	var ids []File
	for _, v := range galleries {
		ids = append(ids, File{
			Size: v.Get("size").Int(),
			Name: v.Get("fileName").String(),
			Id:   v.Get("id").String(),
		})
	}
	return ids, isLastPage, nil
}

//获取文件
func (api *Api) GetPhoto(id string) (io.Reader, error) {
	result, err := api.get(fmt.Sprintf(GetPhoto, time.Now().UnixNano(), id))
	if err != nil {
		return nil, err
	}
	realUrlStr := gjson.Get(string(result), "data.url").String()
	if realUrlStr == "" {
		return nil, errors.New("get fileUrl failed")
	}
	result, err = api.get(realUrlStr)
	if err != nil {
		return nil, err
	}
	realUrl := gjson.Parse(strings.Trim(string(result), "callback()"))
	resp, err := api.User.HttpClient.PostForm(
		realUrl.Get("url").String(),
		url.Values{"meta": []string{realUrl.Get("meta").String()}})
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

// 获取视频相册
func (api *Api) GetVideo() (*Album, error) {
	var ablum *Album
	result, err := api.get(fmt.Sprintf(GetVideo, time.Now().UnixNano()))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(gjson.Get(string(result), "data.album").Raw), &ablum)
	if err != nil {
		return nil, err
	}
	return ablum, nil
}
