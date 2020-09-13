package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/peterh/liner"
	"github.com/thinkoner/openssl"
	"hash"
	"io"
	"os"
	"strconv"
	"time"
)

//md5
func MD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

//sha1
func Sha1(str []byte) string {
	h := sha1.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

//计算文件hash
func FileHash(reader io.Reader, tp string) string {
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
	return fmt.Sprintf("%x", h.Sum(result))
}

func FilePathHash(filePath string, tp string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	return FileHash(file, tp)
}

func AesCBCEncrypt(src, key, iv []byte) (string, error) {
	var dst, err = openssl.AesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(dst), nil
}

func AesCBCDecrypt(dst, key, iv []byte) (string, error) {
	result, err := openssl.AesCBCDecrypt(dst, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

//生成uuid
func GetUuid() string {
	var u uuid.UUID
	var err error
	for i := 0; i < 3; i++ {
		u, err = uuid.NewUUID()
		if err == nil {
			return u.String()
		}
	}
	return ""
}

//生成uuid v4
func GetUuidV4() string {
	var u uuid.UUID
	var err error
	for i := 0; i < 3; i++ {
		u, err = uuid.NewRandom()
		if err == nil {
			return u.String()
		}
	}
	return ""
}

//获取终端输入
func GetInput(msg string) string {
	var newLiner = liner.NewLiner()
	defer newLiner.Close()
	msg = "请输入" + msg + "："
	var prompt = "Go@MiCloud:~$ " + msg
	commandLine, err := newLiner.Prompt(prompt)
	if err != nil {
		fmt.Printf("Prompt Error: %s\n", err)
		os.Exit(-1)
	}
	return commandLine
}

//获取终端输入
func GetInputPwd(msg string) string {
	var newLiner = liner.NewLiner()
	defer newLiner.Close()
	msg = "请输入" + msg + "："
	var prompt = "Go@MiCloud:~$ " + msg
	commandLine, err := newLiner.PasswordPrompt(prompt)
	if err != nil {
		fmt.Printf("Prompt Error: %s\n", err)
		os.Exit(-1)
	}
	return commandLine
}

func FormatTimeInt(t int64, nsec bool) string {
	if nsec {
		i, _ := strconv.Atoi(strconv.Itoa(int(t))[:10])
		t = int64(i)
	}
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}
