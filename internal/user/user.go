package user

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"go-micloud/configs"
	"go-micloud/pkg/utils"
	"go-micloud/pkg/zlog"
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
	UserName     string
	Password     string
	UserId       string
	ServiceToken string
	DeviceId     string
}

func NewUser() *User {
	var jar, _ = cookiejar.New(nil)
	user := User{
		HttpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: zlog.HttpLoggerTransport,
			Jar:       jar,
		},
	}
	go user.autoRenewal()
	return &user
}

func (u *User) autoRenewal() {
	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if u.IsLogin {
				var apiUrl = fmt.Sprintf(autoRenewal, strconv.Itoa(int(time.Now().UnixNano()))[0:13])
				resp, err := u.HttpClient.Get(apiUrl)
				if err != nil {
					zlog.Error(fmt.Sprintf("auto_renewal error: %s", err.Error()))
					return
				}
				if len(resp.Cookies()) > 0 {
					u.updateCookies(imi, resp.Cookies())
				}
			}
		}
	}
}

// 尝试使用保存的Token自动登录
func (u *User) autoLogin() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	var cookies []*http.Cookie
	serviceToken := &http.Cookie{
		Name:   "serviceToken",
		Value:  configs.Conf.ServiceToken,
		Path:   "/",
		Domain: "mi.com",
	}
	userId := &http.Cookie{
		Name:   "userId",
		Value:  configs.Conf.UserId,
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
	zlog.Info(fmt.Sprintf("check_phone_code: %s", result))
	if result == "" {
		u.UserId = configs.Conf.UserId
		u.UserName = configs.Conf.Username
		u.ServiceToken = configs.Conf.ServiceToken
		u.IsLogin = true
		return nil
	} else {
		return errors.New("自动登录失败")
	}
}

// 登录
func (u *User) Login() error {
	err := u.autoLogin()
	if err == nil {
		return nil
	}
	zlog.Info(fmt.Sprintf("auto login failed: %s", err.Error()))
	username, password := u.getConfNamePwd()
	if username == "" || password == "" {
		username, password = u.getInputNamePwd()
	}
	err = u.serviceLogin()
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
	location, err = u.CheckPhoneCode()
	if err != nil {
		return err
	}
	u.UserName = username
	u.Password = password
	// 如果结果不为空，需要手机验证码校验
	if location != "" {
		time.Sleep(time.Millisecond * 500)
		err = u.SendPhoneCode(location)
		if err != nil {
			return err
		}
		err = u.VerifyPhoneCode(utils.GetInput("手机验证码"))
		if err != nil {
			return err
		}
		location, err = u.CheckPhoneCode()
		if err != nil {
			return err
		}
		if location != "" {
			return errors.New("登录失败，请重试！")
		}
	}
	secretPwd, _ := utils.AesCBCEncrypt([]byte(u.Password), []byte("inqH0kEHFvSKqPkR"), []byte("1234567891234500"))
	configs.Conf.Password = secretPwd
	configs.Conf.Username = u.UserName
	configs.Conf.UserId = u.UserId
	configs.Conf.ServiceToken = u.ServiceToken
	configs.Conf.DeviceId = u.DeviceId
	configs.Conf.SaveToFile()
	return nil
}

func (u *User) getConfNamePwd() (string, string) {
	var username = configs.Conf.Username
	var password = configs.Conf.Password
	if password != "" {
		secretPwd, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			password = ""
		} else {
			password, _ = utils.AesCBCDecrypt(secretPwd, []byte("inqH0kEHFvSKqPkR"), []byte("1234567891234500"))
		}
	}
	return username, password
}

func (u *User) getInputNamePwd() (string, string) {
	var username, password string
	for {
		username = utils.GetInput("账号")
		if username == "" {
			zlog.PrintError("账号不能为空")
			continue
		}
		password = utils.GetInput("密码")
		if password == "" {
			zlog.PrintError("密码不能为空")
			continue
		}
		break
	}
	return username, password
}

func (u *User) updateUserInfo(username, password string) error {
	parseUrl, err := url.Parse(drive)
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
		if v.Name == "deviceId" {
			u.DeviceId = v.Value
		}
	}
	u.IsLogin = true
	u.UserName = username
	u.Password = password
	return nil
}

func (u *User) serviceLogin() error {
	request, err := http.NewRequest("GET", serviceLogin, nil)
	if err != nil {
		return err
	}
	deviceId := configs.Conf.DeviceId
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
	}
	resp, err := u.HttpClient.Do(request)
	if err != nil {
		return err
	}
	u.updateCookies(xiaomi, resp.Cookies())
	return nil
}

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
	form.Add("hash", strings.ToUpper(utils.MD5([]byte(password))))
	form.Add("cc", "")
	form.Add("log", `{"title":"dataCenterZone","message":"China"}{"title":"locale","message":"zh_CN"}{"title":"env","message":"release"}{"title":"browser","message":{"name":"Chrome","version":85}}{"title":"search","message":""}{"title":"outerlinkDone","message":"done"}{"title":"addInputChange","message":"userName"}{"title":"loginOrigin","message":"loginMain"}`)
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
		return "", errors.New("登录校验失败")
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
	if resp.StatusCode == http.StatusOK {
		return "", nil
	}
	// 401表示无权限
	if gjson.Get(string(all), "R").Int() == 401 {
		return gjson.Get(string(all), "D").String(), nil
	} else {
		return "", nil
	}
}

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
		return nil
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
	}
	var validCookies []*http.Cookie
	for _, c := range oldCookies {
		if c.Value == "EXPIRED" {
			continue
		}
		validCookies = append(validCookies, c)
		// 更新配置文件
		if c.Name == "userId" {
			u.UserId = c.Value
		}
		if c.Name == "serviceToken" {
			u.ServiceToken = c.Value
		}
		if c.Name == "deviceId" {
			u.DeviceId = c.Value
		}
	}
	jar.SetCookies(parseUrl, validCookies)
	u.HttpClient.Jar = jar
}
