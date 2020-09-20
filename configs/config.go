package configs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var Conf = new(Config)

type Config struct {
	FilePath     string `json:"file_path"`
	UserId       string `json:"user_id"`
	ServiceToken string `json:"service_token"`
	DeviceId     string `json:"device_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	WorkDir      string `json:"work_dir"`
	LogFile      string `json:"log_file"`
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("用户目录不存在: %s\n", err.Error())
	}
	Conf.LogFile = "/tmp/micloud.log"
	Conf.FilePath = userHomeDir + "/.micloud.json"
	// 配置文件不存在
	if _, err := os.Stat(Conf.FilePath); os.IsNotExist(err) {
		Conf.SaveToFile()
	} else {
		file, err := os.Open(Conf.FilePath)
		if err != nil {
			log.Panicf("加载配置文件失败: %s", err.Error())
		}
		all, _ := ioutil.ReadAll(file)
		err = json.Unmarshal(all, &Conf)
		if err != nil {
			log.Panicf("解析配置文件失败: %s", err.Error())
		}
		file.Close()
	}
}

func (r *Config) SaveToFile() {
	file, err := os.Create(r.FilePath)
	if err != nil {
		log.Printf("加载配置文件失败: %s", err.Error())
		return
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Printf("序列化配置文件失败: %s", err.Error())
		return
	}
	_, err = file.WriteString(string(data))
	if err != nil {
		log.Printf("保存配置文件失败: %s", err.Error())
	}
	file.Close()
}
