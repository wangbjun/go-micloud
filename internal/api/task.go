package api

import (
	"fmt"
	"go-micloud/pkg/zlog"
	"time"
)

const (
	TypeDownload = iota
	TypeUpload
)

type Task struct {
	Type         int
	TypeId       string // 上传的父Id|下载的文件Id
	FileName     string
	FilePath     string // 上传的文件路径
	FileSize     int64
	SaveDir      string // 下载保存路径
	Status       int
	StatusMsg    string
	CompleteSize uint64
	Time         time.Time
	RetryTimes   int // 重试次数
}

func (t *Task) getStatusName() string {
	switch t.Status {
	case Waiting:
		return "待处理"
	case Downloading:
		return "下载中"
	case Uploading:
		return "上传中"
	case Succeeded:
		return "已完成"
	case Failed:
		return "失败了"
	default:
		return "未知了"
	}
}

func (t *Task) LogStatus(msg string) {
	t.StatusMsg = msg
	zlog.Info(fmt.Sprintf("[ %s ] %s", t.FileName, msg))
}

func (t *Task) Write(p []byte) (int, error) {
	n := len(p)
	t.CompleteSize += uint64(n)
	return n, nil
}

func (t tasks) Len() int {
	return len(t)
}

func (t tasks) Less(i, j int) bool {
	if t[i].Status == t[j].Status {
		return t[i].FileSize > t[j].FileSize
	}
	return t[i].Status > t[j].Status
}

func (t tasks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
