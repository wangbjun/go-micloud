package configs

import (
	"encoding/json"
	"fmt"
	"go-micloud/pkg/utils"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var Conf *Config

type Config struct {
	FilePath     string `json:"file_path"`
	UserId       string `json:"user_id"`
	ServiceToken string `json:"service_token"`
	DeviceId     string `json:"device_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	WorkDir      string `json:"work_dir"`
	LogFile      string `json:"log_file"`
	UpdatedAt    string `json:"update_at"`
}

func Init(path string) error {
	conf := new(Config)
	conf.FilePath = path
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("打开配置文件失败：%w", err)
		}
		defer file.Close()
		all, err := ioutil.ReadAll(file)
		if err != nil {
			return fmt.Errorf("读取配置文件失败：%w", err)
		}
		err = json.Unmarshal(all, &conf)
		if err != nil {
			return fmt.Errorf("解析配置文件失败：%w", err)
		}
	}
	conf.WorkDir, _ = os.Getwd()
	conf.LogFile = "/tmp/micloud.log"
	conf.SaveToFile()
	Conf = conf
	return nil
}

func (r *Config) SaveToFile() {
	file, err := os.Create(r.FilePath)
	if err != nil {
		log.Printf("创建配置文件失败: %s", err.Error())
		return
	}
	defer file.Close()
	r.UpdatedAt = time.Now().Format(utils.YmdHis)
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Printf("序列化配置文件失败: %s", err.Error())
		return
	}
	_, err = file.WriteString(string(data))
	if err != nil {
		log.Printf("保存配置文件失败: %s", err.Error())
	}
}
