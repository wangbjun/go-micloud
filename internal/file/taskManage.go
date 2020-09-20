package file

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"go-micloud/configs"
	"go-micloud/pkg/zlog"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
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
const DConcurrentNum = 10

type tasks []*task

type TaskManage struct {
	Tasks        tasks
	TotalNum     int64
	DwloadingNum int64
	UploadingNum int64
	Uchan        chan *task
	Dchan        chan *task
	FileApi      *Api
}

func NewManage(fileApi *Api) *TaskManage {
	tm := &TaskManage{
		Tasks:   make(tasks, 0),
		Uchan:   make(chan *task, 1000),
		Dchan:   make(chan *task, 1000),
		FileApi: fileApi,
	}
	go tm.Dispatch()
	return tm
}

func (r *TaskManage) AddDownloadTask(file *File, dir string) {
	task := &task{
		Type:      TypeDownload,
		TypeId:    file.Id,
		FileName:  file.Name,
		FileSize:  file.Size,
		SaveDir:   dir,
		Time:      time.Now(),
		StatusMsg: "等待下载",
	}
	atomic.AddInt64(&r.TotalNum, 1)
	r.Dchan <- task
	r.Tasks = append(r.Tasks, task)
}

func (r *TaskManage) AddUploadTask(fileSize int64, filePath, parentId string) {
	task := &task{
		Type:      TypeUpload,
		TypeId:    parentId,
		FileName:  path.Base(filePath),
		FilePath:  filePath,
		FileSize:  fileSize,
		Time:      time.Now(),
		StatusMsg: "等待上传",
	}
	atomic.AddInt64(&r.TotalNum, 1)
	r.Uchan <- task
	r.Tasks = append(r.Tasks, task)
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
			task.Status = Downloading
			task.StatusMsg = "正在下载"
			go r.download(task)
		}
	}()
}

func (r *TaskManage) download(task *task) {
	defer func() {
		atomic.AddInt64(&r.DwloadingNum, -1)
	}()
	zlog.Info(fmt.Sprintf("开始处理下载任务: %s", task.FileName))
	filePath := configs.Conf.WorkDir + "/" + task.SaveDir + "/" + task.FileName
	if fs, err := os.Stat(filePath); err == nil && fs.Size() == task.FileSize {
		task.LogStatus("文件已存在")
		task.Status = Succeeded
		return
	}
	openFile, err := os.Create(filePath)
	if err != nil {
		r.failed(task, "文件创建失败： "+err.Error())
		return
	}
	reader, err := r.FileApi.GetFile(task.TypeId)
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
	zlog.Info(fmt.Sprintf("文件 [%s] 下载完成，耗时: %f秒", task.FileName, time.Now().Sub(task.Time).Seconds()))
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
	}
	return
}

func (r *TaskManage) ShowTask() {
	if len(r.Tasks) == 0 {
		goto END
	}
	fmt.Printf(strings.Repeat("-", 80) + "\n")
	fmt.Printf("任务状态 |状态信息         |文件总大小 |已处理大小 |文件名\n")
	fmt.Printf(strings.Repeat("-", 80) + "\n")
	sort.Sort(r.Tasks)
	for k, v := range r.Tasks {
		status := v.StatusMsg + strings.Repeat("  ", 8-utf8.RuneCountInString(v.StatusMsg))
		fmt.Printf("%-5s |%s |%-10s |%-10s |%s\n", v.getStatusName(),
			status, humanize.Bytes(uint64(v.FileSize)), humanize.Bytes(v.CompleteSize), v.FileName)
		if k != r.Tasks.Len()-1 && v.Status != r.Tasks[k+1].Status {
			fmt.Printf(strings.Repeat("-", 80) + "\n")
		}
	}
	fmt.Printf(strings.Repeat("-", 80) + "\n")
END:
	fmt.Printf("总任务 %d 个，已完成 %d 个, 待处理 %d 个，处理中 %d 个\n",
		r.TotalNum, int64(len(r.Tasks))-r.DwloadingNum-r.UploadingNum,
		r.TotalNum-int64(len(r.Tasks)), r.DwloadingNum+r.UploadingNum)
}

// 失败重试，最多尝试3次
func (r *TaskManage) failed(task *task, msg string) {
	zlog.Error(fmt.Sprintf("文件 [%s] 下载失败，开始重试，Error: %s", task.FileName, msg))
	task.Time = time.Now()
	task.Status = Failed
	task.StatusMsg = msg
	task.RetryTimes++
	r.Dchan <- task
}
