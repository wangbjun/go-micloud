package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"go-micloud/pkg/utils"
	"go-micloud/pkg/zlog"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const ChunkSize = 4194304

var (
	ErrorSizeTooBig = errors.New("单个文件不能大于4GB")
)

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
func (api *Api) UploadFile(task *task) (string, error) {
	fileInfo, err := os.Stat(task.FilePath)
	if err != nil {
		return "", err
	}
	fileName := path.Base(task.FilePath)
	if fileInfo.Size() >= 4*1024*1024*1024 {
		return "", ErrorSizeTooBig
	}

	task.LogStatus("计算文件哈希值")

	fileSize := fileInfo.Size()
	fileSha1 := utils.FilePathHash(task.FilePath, "sha1")

	task.LogStatus("计算文件分片")
	//大于4MB需要分片
	var blockInfos *[]BlockInfo
	if fileSize > ChunkSize {
		blockInfos, err = api.getFileBlocks(fileInfo, task.FilePath)
		if err != nil {
			return "", errors.New("计算文件分片失败")
		}
	} else {
		blockInfos = &[]BlockInfo{
			{
				Blob: struct {
				}{},
				Sha1: fileSha1,
				Md5:  utils.FilePathHash(task.FilePath, "md5"),
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
	task.StatusMsg = "创建文件分片"
	resp, err := api.postForm(CreateFile, url.Values{
		"data":         []string{string(data)},
		"serviceToken": []string{api.User.ServiceToken},
	})
	if err != nil {
		return "", err
	}
	if result := gjson.Get(string(*resp), "result").String(); result != "ok" {
		zlog.Error("创建文件分片失败：" + string(*resp))
		return "", errors.New("创建文件分片失败")
	}
	isExisted := gjson.Get(string(*resp), "data.storage.exists").Bool()
	//云盘已有此文件
	if isExisted {
		data := UploadJson{Content: UploadContent{
			Name: fileName,
			Storage: UploadExistedStorage{
				UploadId: gjson.Get(string(*resp), "data.storage.uploadId").String(),
				Exists:   true,
			},
		}}
		task.LogStatus("文件已存在")
		return api.createFile(task.TypeId, data)
	} else {
		//云盘不存在该文件
		kss := gjson.Get(string(*resp), "data.storage.kss")
		var (
			nodeUrls   = kss.Get("node_urls").Array()
			fileMeta   = kss.Get("file_meta").String()
			blockMetas = kss.Get("block_metas").Array()
		)
		apiNode := nodeUrls[0].String()
		if apiNode == "" {
			return "", errors.New("暂无可用上传节点")
		}
		//上传分片
		file, err := os.Open(task.FilePath)
		if err != nil {
			return "", err
		}
		var i = 0
		var commitMetas []map[string]string
		for k, block := range blockMetas {
			commitMeta, err := api.uploadBlock(task, k, apiNode, fileMeta, file, block)
			if err != nil {
				return "", err
			}
			commitMetas = append(commitMetas, commitMeta)
			i++
			task.LogStatus(fmt.Sprintf("上传分片(%d/%d)", i, len(*blockInfos)))
		}
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
				UploadId: gjson.Get(string(*resp), "data.storage.uploadId").String(),
				Exists:   false,
			},
		}}
		task.StatusMsg = "上传成功"
		return api.createFile(task.TypeId, data)
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
			Sha1: utils.FileHash(strings.NewReader(string(b)), "sha1"),
			Md5:  utils.FileHash(strings.NewReader(string(b)), "md5"),
			Size: int64(len(b)),
		}
		blockInfos = append(blockInfos, blockInfo)
	}
	return &blockInfos, nil
}

//上传文件分片
func (api *Api) uploadBlock(task *task, num int, apiNode string, fileMeta string, file *os.File, block interface{}) (map[string]string, error) {
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
		task.CompleteSize += uint64(len(fileBlock))
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
		zlog.Error("创建文件失败：" + string(readAll))
		return "", errors.New("创建文件失败")
	} else {
		id := gjson.Get(string(readAll), "data.id").String()
		return id, nil
	}
}

//获取文件公开下载链接
func (api *Api) GetFileDownLoadUrl(id string) (string, error) {
	resp, err := api.User.HttpClient.Get(fmt.Sprintf(GetFiles, id))
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

// 获取目录下的文件
func (api *Api) GetFolder(id string) ([]*File, error) {
	result, err := api.get(fmt.Sprintf(GetFolders, id))
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
		return nil, errors.New("获取文件夹信息失败")
	}
}

func (api *Api) CreateFolder(name, parentId string) (string, error) {
	resp, err := api.postForm(CreateFolder, url.Values{
		"name":         []string{name},
		"parentId":     []string{parentId},
		"serviceToken": []string{api.User.ServiceToken},
	})
	if err != nil {
		return "", err
	}
	if result := gjson.Get(string(*resp), "result").String(); result == "ok" {
		return gjson.Get(string(*resp), "data.id").String(), nil
	} else {
		return "", errors.New("创建目录失败")
	}
}

func (api *Api) DeleteFile(id, fType string) error {
	record := []DeleteFile{{
		Id:   id,
		Type: fType,
	}}
	content, _ := json.Marshal(record)
	resp, err := api.postForm(DeleteFiles, url.Values{
		"operateType":    []string{"DELETE"},
		"operateRecords": []string{string(content)},
		"serviceToken":   []string{api.User.ServiceToken},
	})
	if err != nil {
		return err
	}
	if result := gjson.Get(string(*resp), "result").String(); result == "ok" {
		return nil
	} else {
		return errors.New("删除失败")
	}
}
