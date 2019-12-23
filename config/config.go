package config

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
)

var Conf *ini.File

var EnvFile = "app.ini"

var WorkDir = ""

func init() {
	// 读取配置文件, 解决跑测试的时候找不到配置文件的问题，最多往上找5层目录
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(EnvFile); err == nil {
			break
		} else {
			EnvFile = "../" + EnvFile
		}
	}
	if _, err := os.Stat(EnvFile); os.IsNotExist(err) {
		log.Panicf("conf file [%s]  not found!", EnvFile)
	}
	conf, err := ini.Load(EnvFile)
	if err != nil {
		log.Panicf("parse conf file [%s] failed, err: %s", EnvFile, err.Error())
	}

	Conf = conf

	//工作目录配置
	var workDir = conf.Section("XIAOMI").Key("WORK_DIR").String()
	if workDir != "" {
		WorkDir = workDir
	} else {
		WorkDir, _ = os.Getwd()
	}
}

func SaveToFile() {
	err := Conf.SaveTo(EnvFile)
	if err != nil {
		log.Printf("save config file failed, error %s", err)
	}
}
