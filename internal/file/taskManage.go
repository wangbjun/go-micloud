package file

import (
	"fmt"
	"go-micloud/configs"
	"go-micloud/pkg/zlog"
	"io"
	"os"
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

const (
	TypeDownload = iota
	TypeUpload
)

// 上传并发数
const UConcurrentNum = 5

// 下载并发数
const DConcurrentNum = 10

type task struct {
	File         *File
	Type         int
	ParentId     string // 上传的父Id
	FilePath     string // 上传的文件路径
	SaveDir      string // 下载保存路径
	Status       int
	StatusMsg    string
	CompleteSize uint64
	Time         time.Time
	RetryTimes   int // 重试次数
}

type TaskManage struct {
	Tasks        []*task
	TotalNum     int64
	DwloadingNum int64
	UploadingNum int64
	Uchan        chan *task
	Dchan        chan *task
	FileApi      *Api
}

func NewManage(fileApi *Api) *TaskManage {
	tm := &TaskManage{
		Tasks:   make([]*task, 0),
		Uchan:   make(chan *task, 1000),
		Dchan:   make(chan *task, 1000),
		FileApi: fileApi,
	}
	go tm.Dispatch()
	return tm
}

func (r *TaskManage) AddDownloadTask(file *File, dir string) {
	task := &task{
		Type:    TypeDownload,
		File:    file,
		SaveDir: dir,
		Time:    time.Now(),
		Status:  Waiting,
	}
	atomic.AddInt64(&r.TotalNum, 1)
	r.Dchan <- task
}

func (r *TaskManage) AddUploadTask(filePath, parentId string) {
	task := &task{
		Type:     TypeUpload,
		FilePath: filePath,
		ParentId: parentId,
		Time:     time.Now(),
		Status:   Waiting,
	}
	atomic.AddInt64(&r.TotalNum, 1)
	r.Uchan <- task
}

func (r *TaskManage) Dispatch() {
	go func() {
		for {
			ticker := time.NewTicker(time.Second * 5)
			<-ticker.C
			zlog.Info(fmt.Sprintf("总任务 %d 个，已完成任务 %d 个, 处理中 %d（%d/%d）个",
				r.TotalNum, int64(len(r.Tasks))-r.DwloadingNum-r.UploadingNum,
				r.DwloadingNum+r.UploadingNum, r.UploadingNum, r.DwloadingNum))
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
			atomic.AddInt64(&r.UploadingNum, 1)
			task.Status = Uploading
			go r.upload(task)
			r.Tasks = append(r.Tasks, task)
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
			atomic.AddInt64(&r.DwloadingNum, 1)
			task.Status = Uploading
			go r.download(task)
			r.Tasks = append(r.Tasks, task)
		}
	}()
}

func (r *TaskManage) download(task *task) {
	defer func() {
		atomic.AddInt64(&r.DwloadingNum, -1)
	}()
	zlog.Info(fmt.Sprintf("开始处理下载任务: %s", task.File.Name))
	filePath := configs.WorkDir + "/" + task.SaveDir + "/" + task.File.Name
	if fs, err := os.Stat(filePath); err == nil && fs.Size() == task.File.Size {
		zlog.Info("本地已存在该文件，跳过")
		return
	}
	openFile, err := os.Create(filePath)
	if err != nil {
		r.failed(task, "文件创建失败： "+err.Error())
		return
	}
	reader, err := r.FileApi.GetFile(task.File.Id)
	if err != nil {
		r.failed(task, "文件获取失败： "+err.Error())
		return
	}
	_, err = io.Copy(openFile, io.TeeReader(reader, task))
	if err != nil {
		r.failed(task, "文件写入失败： "+err.Error())
		return
	}
	task.Status = Succeeded
	task.StatusMsg = "下载成功"
	zlog.Info(fmt.Sprintf("文件 [%s] 下载完成，耗时: %f秒", task.File.Name, time.Now().Sub(task.Time).Seconds()))
	return
}

func (r *TaskManage) upload(task *task) {
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
			zlog.Error(fmt.Sprintf("[ %s ]上传失败，重试(%d): %s", task.FilePath, task.RetryTimes, err))
			task.Time = time.Now()
			task.RetryTimes++
			r.Uchan <- task
		}
	} else {
		zlog.Info(fmt.Sprintf("[ %s ]上传成功", task.FilePath))
		task.Status = Succeeded
		task.StatusMsg = "上传成功"
	}
	return
}

// 失败重试，最多尝试3次
func (r *TaskManage) failed(task *task, msg string) {
	zlog.Error(fmt.Sprintf("文件 [%s] 下载失败，开始重试，PrintError: %s", task.File.Name, msg))
	task.Time = time.Now()
	task.Status = Failed
	task.StatusMsg = msg
	task.RetryTimes++
	r.Dchan <- task
}

func (t *task) Write(p []byte) (int, error) {
	n := len(p)
	t.CompleteSize += uint64(n)
	return n, nil
}
