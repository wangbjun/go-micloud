package api

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"go-micloud/pkg/zlog"
	"io"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const (
	Waiting = iota
	Downloading
	Uploading
	Succeeded
	Failed
)

// 上传并发数
const UConcurrentNum = 5

// 下载并发数
const DConcurrentNum = 20

type tasks []*Task

type Manager struct {
	Tasks        tasks
	Num          int64
	DwloadingNum int64
	UploadingNum int64
	Uchan        chan *Task
	Dchan        chan *Task
	FileApi      *Api
}

func NewManager(fileApi *Api) *Manager {
	tm := &Manager{
		Tasks:   make(tasks, 0),
		Uchan:   make(chan *Task, 0),
		Dchan:   make(chan *Task, 0),
		FileApi: fileApi,
	}
	go tm.dispatch()
	return tm
}

func (r *Manager) AddDownloadTask(file *File, saveDir string, tp int) {
	task := &Task{
		Type:      tp,
		TypeId:    file.Id,
		FileName:  file.Name,
		FileSize:  file.Size,
		SaveDir:   saveDir,
		Time:      time.Now(),
		StatusMsg: "等待下载",
	}
	r.Tasks = append(r.Tasks, task)
	r.Dchan <- task
}

func (r *Manager) AddUploadTask(fileSize int64, filePath, parentId string) {
	task := &Task{
		Type:      TypeUpload,
		TypeId:    parentId,
		FileName:  path.Base(filePath),
		FilePath:  filePath,
		FileSize:  fileSize,
		Time:      time.Now(),
		StatusMsg: "等待上传",
	}
	r.Tasks = append(r.Tasks, task)
	r.Uchan <- task
}

func (r *Manager) dispatch() {
	go func() {
		for {
			ticker := time.NewTicker(time.Second * 5)
			<-ticker.C
			zlog.Info(fmt.Sprintf("总任务 %d 个，已完成 %d 个, 待处理 %d 个，处理中 %d 个\n",
				len(r.Tasks), r.Num-r.DwloadingNum-r.UploadingNum,
				int64(len(r.Tasks))-r.Num, r.DwloadingNum+r.UploadingNum))
		}
	}()
	go func() {
		for {
			if r.UploadingNum >= UConcurrentNum {
				time.Sleep(time.Second)
				continue
			}
			task := <-r.Uchan
			if task == nil || task.RetryTimes > 3 {
				continue
			}
			if task.RetryTimes == 0 {
				atomic.AddInt64(&r.Num, 1)
			}
			atomic.AddInt64(&r.UploadingNum, 1)
			task.Status = Uploading
			go r.upload(task)
		}
	}()
	go func() {
		for {
			if r.DwloadingNum >= DConcurrentNum {
				time.Sleep(time.Second)
				continue
			}
			task := <-r.Dchan
			if task == nil || task.RetryTimes > 3 {
				continue
			}
			if task.RetryTimes == 0 {
				atomic.AddInt64(&r.Num, 1)
			}
			atomic.AddInt64(&r.DwloadingNum, 1)
			task.Status = Downloading
			task.StatusMsg = "正在下载"
			go r.download(task)
		}
	}()
}

func (r *Manager) download(task *Task) {
	defer func() {
		atomic.AddInt64(&r.DwloadingNum, -1)
	}()
	zlog.Info(fmt.Sprintf("开始处理下载任务: %s", task.FileName))
	filePath := task.SaveDir + "/" + task.FileName
	if fs, err := os.Stat(filePath); err == nil && fs.Size() == task.FileSize {
		task.LogStatus("文件已存在")
		task.Status = Succeeded
		return
	}
	var err error
	var reader io.ReadCloser
	if task.Type == TypeDownload {
		reader, err = r.FileApi.GetFile(task.TypeId)
	} else if task.Type == TypeDownloadAlbum {
		reader, err = r.FileApi.GetPhoto(task.TypeId)
	}
	if err != nil {
		go r.failed(task, "文件获取失败： "+err.Error())
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		go r.failed(task, "文件创建失败： "+err.Error())
		return
	}
	defer file.Close()
	defer reader.Close()
	_, err = io.Copy(file, io.TeeReader(reader, task))
	if err != nil {
		go r.failed(task, "文件写入失败： "+err.Error())
		return
	}
	task.Status = Succeeded
	task.StatusMsg = "下载成功"
	zlog.Info(fmt.Sprintf("文件 [%s] 下载完成，耗时: %f秒", task.FileName, time.Now().Sub(task.Time).Seconds()))
	return
}

func (r *Manager) upload(task *Task) {
	defer func() {
		atomic.AddInt64(&r.UploadingNum, -1)
	}()
	zlog.Info(fmt.Sprintf("开始处理上传任务: %s", task.FilePath))
	_, err := r.FileApi.UploadFile(task)
	if err != nil {
		task.Status = Failed
		task.StatusMsg = err.Error()
		if err == ErrorSizeTooBig {
			zlog.Error(fmt.Sprintf("[ %s ]上传失败: %s", task.FilePath, err))
		} else {
			go func() {
				time.Sleep(time.Second * 30)
				zlog.Error(fmt.Sprintf("[ %s ]上传失败，重试(%d): %s", task.FilePath, task.RetryTimes, err))
				task.Time = time.Now()
				task.RetryTimes++
				r.Uchan <- task
			}()
		}
	} else {
		zlog.Info(fmt.Sprintf("[ %s ]上传成功", task.FilePath))
		task.Status = Succeeded
	}
	return
}

func (r *Manager) ShowTask() {
	var i = 1
	if len(r.Tasks) == 0 {
		goto END
	}
	fmt.Printf(strings.Repeat("-", 80) + "\n")
	fmt.Printf(" 序号 | 任务状态 | 状态信息 | 文件总大小 | 已处理大小 | 文件名\n")
	fmt.Printf(strings.Repeat("-", 80) + "\n")
	sort.Sort(r.Tasks)
	for k, v := range r.Tasks {
		fmt.Printf(" %-3d| %-3s | %s | %-7s | %-7s | %s\n", i, v.getStatusName(),
			v.StatusMsg, humanize.Bytes(uint64(v.FileSize)), humanize.Bytes(v.CompleteSize), v.SaveDir+"/"+v.FileName)
		if k != r.Tasks.Len()-1 && v.Status != r.Tasks[k+1].Status {
			i = 0
			fmt.Printf(strings.Repeat("-", 80) + "\n")
		}
		i++
	}
	fmt.Printf(strings.Repeat("-", 80) + "\n")
END:
	fmt.Printf("总任务 %d 个，已完成 %d 个, 待处理 %d 个，处理中 %d 个\n",
		len(r.Tasks), r.Num-r.DwloadingNum-r.UploadingNum,
		int64(len(r.Tasks))-r.Num, r.DwloadingNum+r.UploadingNum)
}

// 失败重试，最多尝试3次
func (r *Manager) failed(task *Task, msg string) {
	time.Sleep(time.Second * time.Duration(rand.Intn(60)+60))
	task.Time = time.Now()
	task.Status = Failed
	task.StatusMsg = msg
	task.RetryTimes++
	zlog.Error(fmt.Sprintf("文件 [%s] 下载失败，开始重试(%d)，Error: %s", task.FileName, task.RetryTimes, msg))
	r.Dchan <- task
}
