package api

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/tidwall/gjson"
	"go-micloud/pkg/color"
	"go-micloud/pkg/zlog"
	"hash"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	BaseUri     = "https://i.mi.com"
	GetFiles    = BaseUri + "/drive/user/files/%s?jsonpCallback=callback"
	CreateFile  = BaseUri + "/drive/user/files/create"
	UploadFile  = BaseUri + "/drive/user/files"
	DeleteFiles = BaseUri + "/drive/user/files/%s/del"
)

const ChunkSize = 4194304

//获取文件公开下载链接
func (api *Api) GetFileDownLoadUrl(id string) (string, error) {
	var apiUrl = fmt.Sprintf(GetFiles, id)
	resp, err := api.User.HttpClient.Get(apiUrl)
	if err != nil {
		return "", err
	}
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	realUrlStr := gjson.Get(string(all), "data.storage.jsonpUrl").String()
	if realUrlStr == "" {
		return "", errors.New("get fileUrl failed")
	}
	result, err := api.get(realUrlStr)
	if err != nil {
		return "", err
	}
	realUrl := gjson.Parse(strings.Trim(string(result), "callback()"))
	return realUrl.String(), nil
}

//获取文件
func (api *Api) GetFile(id string) (io.Reader, error) {
	result, err := api.get(fmt.Sprintf(GetFiles, id))
	if err != nil {
		return nil, err
	}
	realUrlStr := gjson.Get(string(result), "data.storage.jsonpUrl").String()
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

//上传文件
func (api *Api) UploadFile(filePath string, parentId string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	fileName := path.Base(filePath)
	zlog.Info(fmt.Sprintf("文件大小: %s", humanize.Bytes(uint64(fileInfo.Size()))))

	if fileInfo.Size() == 0 || fileInfo.Size() >= 4*1024*1024*1024 {
		return "", errors.New("目前不支持大于4GB文件上传")
	}
	zlog.Info("计算文件sha1")
	fileSize := fileInfo.Size()
	fileSha1 := calFileHash(filePath, "sha1")

	var blockInfos *[]BlockInfo
	//大于4MB需要分片

	zlog.Info("计算文件分片信息")

	if fileSize > ChunkSize {
		blockInfos, err = api.getFileBlocks(fileInfo, filePath)
		if err != nil {
			return "", errors.New("get file blocks failed")
		}
	} else {
		blockInfos = &[]BlockInfo{
			{
				Blob: struct {
				}{},
				Sha1: fileSha1,
				Md5:  calFileHash(filePath, "md5"),
				Size: fileSize,
			},
		}
	}
	var uploadJson = UploadJson{
		Content: UploadContent{
			Name: fileName,
			Storage: UploadStorage{
				Size: fileSize,
				Sha1: fileSha1,
				Kss: UploadKss{
					BlockInfos: *blockInfos,
				},
			},
		},
	}
	data, _ := json.Marshal(uploadJson)

	//创建分片
	zlog.Info(fmt.Sprintf("创建文件分片(%d)", len(*blockInfos)))

	resp, err := api.User.HttpClient.PostForm(CreateFile, url.Values{
		"data":         []string{string(data)},
		"serviceToken": []string{api.User.ServiceToken},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	all, _ := ioutil.ReadAll(resp.Body)
	if result := gjson.Get(string(all), "result").String(); result != "ok" {
		return "", errors.New("create file failed, error: " + gjson.Get(string(all), "description").String())
	}
	isExisted := gjson.Get(string(all), "data.storage.exists").Bool()
	//云盘已有此文件
	if isExisted {
		data := UploadJson{Content: UploadContent{
			Name: fileName,
			Storage: UploadExistedStorage{
				UploadId: gjson.Get(string(all), "data.storage.uploadId").String(),
				Exists:   true,
			},
		}}
		zlog.Info("当前文件已存在,上传完成")
		return api.createFile(parentId, data)
	} else {
		//云盘不存在该文件
		kss := gjson.Get(string(all), "data.storage.kss")
		var (
			nodeUrls   = kss.Get("node_urls").Array()
			fileMeta   = kss.Get("file_meta").String()
			blockMetas = kss.Get("block_metas").Array()
		)
		apiNode := nodeUrls[0].String()
		if apiNode == "" {
			return "", errors.New("no available url node")
		}
		//上传分片
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		var i = 0
		var commitMetas []map[string]string
		for k, block := range blockMetas {
			commitMeta, err := api.uploadBlock(k, apiNode, fileMeta, file, block)
			if err != nil {
				return "", err
			}
			commitMetas = append(commitMetas, commitMeta)
			i++
			fmt.Printf("\r%s", strings.Repeat(" ", 35))
			fmt.Printf("\r" + color.Green(fmt.Sprintf("### Info: 正在上传分片(%d/%d)", i, len(*blockInfos))))
		}
		fmt.Printf("\n")
		//最终完成上传
		data := UploadJson{Content: UploadContent{
			Name: fileName,
			Storage: UploadStorage{
				Size: fileSize,
				Sha1: fileSha1,
				Kss: Kss{
					Stat:            "OK",
					NodeUrls:        nodeUrls,
					SecureKey:       kss.Get("secure_key").String(),
					ContentCacheKey: kss.Get("contentCacheKey").String(),
					FileMeta:        kss.Get("file_meta").String(),
					CommitMetas:     commitMetas,
				},
				UploadId: gjson.Get(string(all), "data.storage.uploadId").String(),
				Exists:   false,
			},
		}}
		zlog.Info("所有分片上传完毕，上传完成")
		return api.createFile(parentId, data)
	}
}

//获取文件分片信息
func (api *Api) getFileBlocks(fileInfo os.FileInfo, filePath string) (*[]BlockInfo, error) {
	num := int(math.Ceil(float64(fileInfo.Size()) / float64(ChunkSize)))
	file, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	var i int64 = 1
	var blockInfos []BlockInfo
	for b := make([]byte, ChunkSize); i <= int64(num); i++ {
		offset := (i - 1) * ChunkSize
		_, _ = file.Seek(offset, 0)
		if len(b) > int(fileInfo.Size()-offset) {
			b = make([]byte, fileInfo.Size()-offset)
		}
		_, err := file.Read(b)
		if err != nil {
			continue
		}
		blockInfo := BlockInfo{
			Blob: struct{}{},
			Sha1: calHash(strings.NewReader(string(b)), "sha1"),
			Md5:  calHash(strings.NewReader(string(b)), "md5"),
			Size: int64(len(b)),
		}
		blockInfos = append(blockInfos, blockInfo)
	}
	return &blockInfos, nil
}

//上传文件分片
func (api *Api) uploadBlock(num int, apiNode string, fileMeta string, file *os.File, block interface{}) (map[string]string, error) {
	m, ok := (block).(gjson.Result)
	if !ok {
		return nil, errors.New("block info error")
	}
	//block已存在则不上传
	if m.Get("is_existed").Int() == 1 {
		return map[string]string{"commit_meta": m.Get("commit_meta").String()}, nil
	} else {
		uploadUrl := apiNode + "/upload_block_chunk?chunk_pos=0&file_meta=" + fileMeta + "&block_meta=" + m.Get("block_meta").String()
		fileInfo, _ := file.Stat()
		offset := int64(num * ChunkSize)
		chunkSize := ChunkSize
		if chunkSize > int(fileInfo.Size()-offset) {
			chunkSize = int(fileInfo.Size() - offset)
		}
		fileBlock := make([]byte, chunkSize)
		_, err := file.Seek(offset, 0)
		_, err = file.Read(fileBlock)
		if err != nil {
			return nil, err
		}
		request, _ := http.NewRequest("POST", uploadUrl, strings.NewReader(string(fileBlock)))
		request.Header.Set("DNT", "1")
		request.Header.Set("Origin", "https://i.mi.com")
		request.Header.Set("Referer", "https://i.mi.com/drive")
		request.Header.Set("Content-Type", "application/octet-stream")
		response, err := api.User.HttpClient.Do(request)
		if err != nil {
			return nil, err
		}
		readAll, err := ioutil.ReadAll(response.Body)
		stat := gjson.Get(string(readAll), "stat").String()
		if stat != "BLOCK_COMPLETED" {
			return nil, errors.New("block not completed")
		}
		response.Body.Close()
		return map[string]string{"commit_meta": gjson.Get(string(readAll), "commit_meta").String()}, nil
	}
}

//最终创建文件
func (api *Api) createFile(parentId string, data interface{}) (string, error) {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Add("data", string(dataJson))
	form.Add("serviceToken", api.User.ServiceToken)
	form.Add("parentId", parentId)
	request, _ := http.NewRequest("POST", UploadFile, strings.NewReader(form.Encode()))
	request.Header.Set("DNT", "1")
	request.Header.Set("Origin", "https://i.mi.com")
	request.Header.Set("Referer", "https://i.mi.com/drive")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := api.User.HttpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	readAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if result := gjson.Get(string(readAll), "result").String(); result != "ok" {
		return "", errors.New(gjson.Get(string(readAll), "description").String())
	} else {
		id := gjson.Get(string(readAll), "data.id").String()
		return id, nil
	}
}

func (api *Api) get(url string) ([]byte, error) {
	result, err := api.User.HttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if result.StatusCode == http.StatusFound {
		result, err = api.User.HttpClient.Get(result.Header.Get("Location"))
		if err != nil {
			return nil, err
		}
	}
	bytes, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	return bytes, nil
}

func calFileHash(filePath string, tp string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	return calHash(file, tp)
}

func calHash(reader io.Reader, tp string) string {
	var result []byte
	var h hash.Hash
	if tp == "md5" {
		h = md5.New()
	} else {
		h = sha1.New()
	}
	if _, err := io.Copy(h, reader); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(result))
}
