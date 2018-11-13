package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	ServerHost = "https://csust.edu.chsh8j.com"
	ClientName = "Huawei Mate 20 Pro 中国联通 7.1.1"
)

var (
	username = flag.String("u", "", "学号/身份证号/手机号")
	password = flag.String("p", "", "密码")
	bleId    = flag.String("b", "", "蓝牙ID")
	devUUID  = flag.String("d", "", "设备UUID")
)

type SignBot struct {
	name     string
	token    string
	signId   string
	signTime string
}

type SignBotError struct {
	message string
}

func NewSignBot() *SignBot {
	return &SignBot{}
}

func (sbe *SignBotError) Error() string {
	return sbe.message
}
func (sb *SignBot) Login(username string, password string) error {
	jsonByte, err := json.Marshal(map[string]string{"userName": username, "password": password})
	if err != nil {
		return err
	}
	resp, err := http.PostForm(ServerHost+"/magus/appuserloginapi/userlogin", url.Values{"params": {string(jsonByte)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result struct {
		Result struct {
			IsSuccess string `json:"isSuccess"`
			Message   string `json:"message"`
			Token     string `json:"token"`
			Userinfo  struct {
				UserZhname string `json:"userZhname"`
			} `json:"userinfo"`
		} `json:"result"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	if result.Result.IsSuccess != "1" {
		return &SignBotError{result.Result.Message}
	}
	sb.token = result.Result.Token
	sb.name = result.Result.Userinfo.UserZhname
	return nil
}

func (sb *SignBot) Detail() error {
	req, err := http.NewRequest("POST", ServerHost+"/dorm/app/dormsign/sign/student/detail",
		strings.NewReader(url.Values{"token": {sb.token}}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("token", sb.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result struct {
		Data struct {
			IsAvailable string `json:"isAvailable"`
			SignId      string `json:"signId"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	if result.Data.IsAvailable != "1" {
		return &SignBotError{"已签到过或今天没有签到任务"}
	}
	sb.signId = result.Data.SignId
	return nil
}
func (sb *SignBot) Sign(bleId string, devUUID string) error {
	req, err := http.NewRequest("POST", ServerHost+"/dorm/app/dormsign/sign/student/edit",
		strings.NewReader(url.Values{
			"signId":  {sb.signId},
			"bleId":   {bleId},
			"devUuid": {devUUID},
			"osName":  {ClientName},
		}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("token", sb.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result struct {
		Data struct {
			IsSuccess string `json:"isSuccess"`
			Message   string `json:"message"`
			SignTime  string `json:"signTime"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	if result.Data.IsSuccess != "1" {
		return &SignBotError{result.Data.Message}
	}
	sb.signTime = result.Data.SignTime
	return nil
}

func main() {
	flag.Parse()
	signBot := NewSignBot()
	err := signBot.Login(*username, *password)
	if err != nil {
		fmt.Printf("登录失败：%s\n", err.Error())
		os.Exit(1)
	}
	err = signBot.Detail()
	if err != nil {
		fmt.Printf("签到失败：%s\n", err.Error())
		os.Exit(1)
	}
	err = signBot.Sign(*bleId, *devUUID)
	if err != nil {
		fmt.Printf("签到失败：%s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("[签到成功]姓名：%s，签到时间：%s\n", signBot.name, signBot.signTime)
	os.Exit(0)
}
