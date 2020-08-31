package configs

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
)

var Conf *ini.File

const configTpl = `[XIAOMI]
USER_ID =

SERVICE_TOKEN =

WORK_DIR =

DEVICE_ID =

[XIAOMI_ACCOUNT]
USERNAME =

PASSWORD =

[APP]
LOG_FILE = /tmp/micloud.log
`

var EnvFile = ".micloud.ini"

var WorkDir = ""

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("用户目录不存在, err: %s\n", err.Error())
	}
	EnvFile = userHomeDir + "/" + EnvFile
	if _, err := os.Stat(EnvFile); os.IsNotExist(err) {
		file, err := os.Create(EnvFile)
		if err != nil {
			log.Panicln(err.Error())
		}
		_, err = file.WriteString(configTpl)
		if err != nil {
			log.Panicf("init config file failed, err: %s\n", err.Error())
		}
		file.Sync()
		file.Close()
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
