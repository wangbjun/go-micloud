package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"go-micloud/config"
	"go-micloud/lib/function"
	"go-micloud/lib/zlog"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	xiaomi           = "https://account.xiaomi.com"
	imi              = "https://i.mi.com"
	drive            = imi + "/drive/user/folders/0/children"
	autoRenewal      = imi + "/status/setting?type=AutoRenewal&inactiveTime=10&_dc=%s"
	serviceLogin     = xiaomi + "/pass/serviceLogin"
	serviceLoginAuth = xiaomi + "/pass/serviceLoginAuth2?_dc=%s"
	sendPhoneCode    = xiaomi + "/auth/sendPhoneTicket?_dc=%s"
	verifyPhoneCode  = xiaomi + "/auth/verifyPhone?_flag=4?_dc=%s"
)

type User struct {
	HttpClient   *http.Client
	IsLogin      bool
	UserId       string
	ServiceToken string
}

var Account *User

func init() {
	Account = NewUser()
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ticker.C:
				if Account.IsLogin {
					err := Account.autoRenewal()
					if err != nil {
						fmt.Printf("autoRenewal error: %s", err)
					}
				}
			}
		}
	}()
}

func NewUser() *User {
	var jar, _ = cookiejar.New(nil)
	return &User{
		HttpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: zlog.HttpLoggerTransport,
			Jar:       jar,
		},
	}
}

func (u *User) autoRenewal() error {
	var apiUrl = fmt.Sprintf(autoRenewal, strconv.Itoa(int(time.Now().UnixNano()))[0:13])
	resp, err := u.HttpClient.Get(apiUrl)
	if err != nil {
		return err
	}
	if len(resp.Cookies()) > 0 {
		u.updateCookies(imi, resp.Cookies())
	}
	return nil
}

// 手动录入cookies登录
func (u *User) LoginManual() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	var cookies []*http.Cookie
	var (
		cServiceToken = config.Conf.Section("XIAOMI").Key("SERVICE_TOKEN").String()
		cUserId       = config.Conf.Section("XIAOMI").Key("USER_ID").String()
	)
	serviceToken := &http.Cookie{
		Name:   "serviceToken",
		Value:  cServiceToken,
		Path:   "/",
		Domain: "mi.com",
	}
	userId := &http.Cookie{
		Name:   "userId",
		Value:  cUserId,
		Path:   "/",
		Domain: "mi.com",
	}
	cookies = append(cookies, serviceToken, userId)
	parseUrl := &url.URL{
		Scheme: "https",
		Host:   "mi.com",
		Path:   "/",
	}
	jar.SetCookies(parseUrl, cookies)
	u.HttpClient.Jar = jar

	result, err := u.CheckPhoneCode()
	if err != nil {
		return err
	}
	if result == "" {
		u.UserId = cUserId
		u.ServiceToken = cServiceToken
		u.IsLogin = true
		return nil
	} else {
		return errors.New("登录失败，请重试")
	}
}

func (u *User) Login(input bool) error {
	var username string
	var password string
	var account = config.Conf.Section("XIAOMI_ACCOUNT")
	for {
		if account.Key("USERNAME").String() != "" && !input {
			username = account.Key("USERNAME").String()
		} else {
			username = function.GetInput("账号")
			if username == "" {
				fmt.Println("===> 账号不能为空")
				continue
			}
		}
	PASS:
		if account.Key("PASSWORD").String() != "" && !input {
			secretPwd, _ := base64.StdEncoding.DecodeString(account.Key("PASSWORD").String())
			password, _ = function.AesCBCDecrypt(secretPwd,
				[]byte("inqH0kEHFvSKqPkR"), []byte("1234567891234500"))
		} else {
			password = function.GetInputPwd("密码")
			if len(password) < 6 {
				fmt.Println("===> 密码不能少于6位")
				goto PASS
			}
		}
		break
	}
	err := u.serviceLogin()
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 500)
	location, err := u.serviceLoginAuth(username, password)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 500)
	err = u.passServiceLogin(location)
	if err != nil {
		return err
	}
	parseUrl, err := url.Parse(serviceLogin)
	if err != nil {
		return err
	}
	cookies := u.HttpClient.Jar.Cookies(parseUrl)
	for _, v := range cookies {
		if v.Name == "userId" {
			u.UserId = v.Value
		}
		if v.Name == "serviceToken" {
			u.ServiceToken = v.Value
		}
	}
	result, err := u.CheckPhoneCode()
	if err != nil {
		return err
	}
	if result == "" {
		u.IsLogin = true
		return nil
	}
	err = u.SendPhoneCode(result)
	if err == ErrorNotNeedSms {
		u.IsLogin = true
		fmt.Println("===> 登录成功！")
		go saveAccount(username, password)
		return nil
	} else {
		if err != nil {
			return err
		}
		err = u.VerifyPhoneCode(function.GetInput("手机验证码"))
		if err != nil {
			return err
		}
		result, err = u.CheckPhoneCode()
		if err != nil {
			return err
		}
		if result == "" {
			u.IsLogin = true
			fmt.Println("===> 登录成功！")
			go saveAccount(username, password)
			return nil
		} else {
			return errors.New("登录失败，请重试！")
		}
	}
}

func (u *User) serviceLogin() error {
	request, err := http.NewRequest("GET", serviceLogin, nil)
	if err != nil {
		return err
	}
	var (
		deviceId = config.Conf.Section("XIAOMI").Key("DEVICE_ID").String()
	)
	if deviceId != "" {
		var cookies []*http.Cookie
		cookies = append(cookies, &http.Cookie{
			Name:    "deviceId",
			Value:   deviceId,
			Path:    "/",
			Domain:  "xiaomi.com",
			Expires: time.Now().AddDate(50, 0, 0),
		}, &http.Cookie{
			Name:    "pass_ua",
			Value:   "web",
			Path:    "/",
			Domain:  "xiaomi.com",
			Expires: time.Now().AddDate(50, 0, 0),
		})
		u.updateCookies(xiaomi, cookies)
	}
	resp, err := u.HttpClient.Do(request)
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())
	return nil
}

var ErrorPwd = errors.New("登录校验失败")

func (u *User) serviceLoginAuth(userName, password string) (string, error) {
	var apiUrl = fmt.Sprintf(serviceLoginAuth, strconv.Itoa(int(time.Now().UnixNano()))[0:13])
	form := url.Values{}
	form.Add("_json", "true")
	form.Add("callback", xiaomi)
	form.Add("sid", "passport")
	form.Add("qs", "%3Fsid%3Dpassport")
	form.Add("_sign", "2&V1_passport&wqS4omyjALxMm//3wLXcVcITjEc=")
	form.Add("serviceParam", `{"checkSafePhone":false}`)
	form.Add("user", userName)
	form.Add("hash", strings.ToUpper(function.MD5([]byte(password))))
	form.Add("cc", "")
	form.Add("log", `{"title":"dataCenterZone","message":"China"}{"title":"locale","message":"zh_CN"}{"title":"env","message":"release"}{"title":"browser","message":{"name":"Chrome","version":79}}{"title":"search","message":""}{"title":"outerlinkDone","message":"done"}{"title":"addInputChange","message":"userName"}{"title":"loginOrigin","message":"loginMain"}`)
	resp, err := u.HttpClient.PostForm(apiUrl, form)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	u.updateCookies(xiaomi, resp.Cookies())

	all = []byte(strings.Trim(string(all), "&&&START&&&"))

	location := gjson.Get(string(all), "location").String()
	if location == "" {
		return "", ErrorPwd
	}

	return location, nil
}

func (u *User) passServiceLogin(location string) error {
	resp, err := u.HttpClient.Get(location)
	if err != nil {
		return err
	}
	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}
	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())

	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}

	u.updateCookies(xiaomi, resp.Cookies())

	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}
	return nil
}

func (u *User) CheckPhoneCode() (string, error) {
	resp, err := u.HttpClient.Get(drive)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// 401表示无权限
	if gjson.Get(string(all), "R").Int() == 401 {
		return gjson.Get(string(all), "D").String(), nil
	} else {
		return "", nil
	}
}

var ErrorNotNeedSms = errors.New("not need SMS")

func (u *User) SendPhoneCode(location string) error {
	resp, err := u.HttpClient.Get(location)
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())

	location = resp.Header.Get("Location")
	resp, err = u.HttpClient.Get(location)
	if err != nil {
		return err
	}
	if strings.Contains(location, "i.mi.com") {
		u.updateCookies(imi, resp.Cookies())
		return ErrorNotNeedSms
	}
	u.updateCookies(xiaomi, resp.Cookies())

	apiUrl := fmt.Sprintf(sendPhoneCode, strconv.Itoa(int(time.Now().UnixNano()))[0:13])
	form := url.Values{}
	form.Add("user", u.UserId)
	form.Add("retry", "0")
	resp, err = u.HttpClient.PostForm(apiUrl, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	u.updateCookies(xiaomi, resp.Cookies())

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	all = []byte(strings.Trim(string(all), "&&&START&&&"))
	result := gjson.Get(string(all), "result").String()
	if result != "ok" {
		if gjson.Get(string(all), "code").Int() == 70022 {
			return errors.New("该账号今日手机验证码次数超出限制(5次)，请24小时后再试！")
		}
		return errors.New(gjson.Get(string(all), "description").String())
	}

	return nil
}

func (u *User) VerifyPhoneCode(code string) error {
	var apiUrl = fmt.Sprintf(verifyPhoneCode, strconv.Itoa(int(time.Now().UnixNano()))[0:13])
	form := url.Values{}
	form.Add("_json", "true")
	form.Add("user", u.UserId)
	form.Add("ticket", code)
	form.Add("trust", "true")
	request, err := http.NewRequest("POST", apiUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("Content-type", "application/x-www-form-urlencoded")
	resp, err := u.HttpClient.Do(request)
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())

	defer resp.Body.Close()
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	location := gjson.Get(string(all), "location").String()
	if location == "" {
		return errors.New("验证码错误")
	}

	resp, err = u.HttpClient.Get(xiaomi + location)
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())

	defer resp.Body.Close()
	all, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}
	u.updateCookies(imi, resp.Cookies())

	defer resp.Body.Close()
	all, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp, err = u.HttpClient.Get(resp.Header.Get("Location"))
	if err != nil {
		return err
	}

	return nil
}

func (u *User) updateCookies(domain string, newCookies []*http.Cookie) {
	jar, _ := cookiejar.New(nil)
	parseUrl, _ := url.Parse(domain)
	var oldCookies = u.HttpClient.Jar.Cookies(parseUrl)
	for _, v := range newCookies {
		var existed = false
		for _, k := range oldCookies {
			if k.Name == v.Name {
				k = v
				existed = true
			}
		}
		if !existed {
			oldCookies = append(oldCookies, v)
		}
		if v.Name == "deviceId" {
			go saveDeviceId(v)
		}
	}
	var validCookies []*http.Cookie
	for _, c := range oldCookies {
		if c.Value != "EXPIRED" {
			validCookies = append(validCookies, c)
		}
		// 更新配置文件
		if c.Name == "userId" && c.Value != u.UserId {
			config.Conf.Section("XIAOMI").Key("USER_ID").SetValue(c.Value)
			go config.SaveToFile()
		}
		if c.Name == "serviceToken" && c.Value != u.ServiceToken {
			config.Conf.Section("XIAOMI").Key("SERVICE_TOKEN").SetValue(c.Value)
			go config.SaveToFile()
		}
	}
	jar.SetCookies(parseUrl, validCookies)
	u.HttpClient.Jar = jar
}

func saveDeviceId(v *http.Cookie) {
	deviceKey := config.Conf.Section("XIAOMI").Key("DEVICE_ID")
	deviceId := deviceKey.String()
	if deviceId == "" {
		deviceKey.SetValue(v.Value)
	}
}

func saveAccount(username, password string) {
	account := config.Conf.Section("XIAOMI_ACCOUNT")
	account.Key("USERNAME").SetValue(username)

	secretPwd, _ := function.AesCBCEncrypt([]byte(password),
		[]byte("inqH0kEHFvSKqPkR"), []byte("1234567891234500"))

	account.Key("PASSWORD").SetValue(secretPwd)
	go config.SaveToFile()
}
