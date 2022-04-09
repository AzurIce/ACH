package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type PlayerInfo struct {
	UUID string
	Name string
}

func GetPlayerInfo(mcToken string) PlayerInfo {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: tr,
	}

	// fmt.Println(string(jsonStr))

	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+mcToken)

	res, err := client.Do(req)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Panicf("获取player info: 请求失败, %s\n", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("获取player info: 读取Body失败, %s\n", err)
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("获取player info: json解析失败, %s\n", err)
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	return PlayerInfo{
		data["id"].(string),
		data["name"].(string),
	}
}

func GetMCToken(xsts XSTSTokenResponse) string {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: tr,
	}

	reqBody := make(map[string]interface{})
	reqBody["identityToken"] = fmt.Sprintf("XBL3.0 x=%s;%s", xsts.Uhs, xsts.Token)

	jsonStr, _ := json.Marshal(reqBody)

	// fmt.Println(string(jsonStr))

	req, _ := http.NewRequest("POST", "https://api.minecraftservices.com/authentication/login_with_xbox", bytes.NewReader(jsonStr))

	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Panicf("获取mc token: 请求失败, %s\n", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("获取mc token: 读取Body失败, %s\n", err)
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("获取mc token: json解析失败, %s\n", err)
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	return data["access_token"].(string)
}

func XSTSTokenRequestBody(xblToken string) map[string]interface{} {
	data := make(map[string]interface{})
	data2 := make(map[string]interface{})
	data2["SandboxId"] = "RETAIL"
	data2["UserTokens"] = []string{xblToken}
	data["Properties"] = data2
	data["RelyingParty"] = "rp://api.minecraftservices.com/"
	data["TokenType"] = "JWT"
	return data
}

type XSTSTokenResponse struct {
	Token string
	Uhs   string
}

func GetXSTSToken(xblToken string) XSTSTokenResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: tr,
	}

	jsonStr, _ := json.Marshal(XSTSTokenRequestBody(xblToken))

	// fmt.Println(string(jsonStr))

	req, _ := http.NewRequest("POST", "https://xsts.auth.xboxlive.com/xsts/authorize", bytes.NewReader(jsonStr))

	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Panicf("获取xsts token: 请求失败, %s\n", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("获取xsts token: 读取Body失败, %s\n", err)
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("获取xsts token: json解析失败, %s\n", err)
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	return XSTSTokenResponse{
		data["Token"].(string),
		data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string),
	}
}

type XBLTokenResponse struct {
	Token string
	Uhs   string
}

func XBLTokenRequestBody(token string) map[string]interface{} {
	data := make(map[string]interface{})
	data2 := make(map[string]interface{})
	data2["AuthMethod"] = "RPS"
	data2["SiteName"] = "user.auth.xboxlive.com"
	data2["RpsTicket"] = "d=" + token
	data["Properties"] = data2
	data["RelyingParty"] = "http://auth.xboxlive.com"
	data["TokenType"] = "JWT"
	return data
}

func GetXBLToken(token string) XBLTokenResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: tr,
	}

	jsonStr, _ := json.Marshal(XBLTokenRequestBody(token))

	// fmt.Println(string(jsonStr))

	req, _ := http.NewRequest("POST", "https://user.auth.xboxlive.com/user/authenticate", bytes.NewReader(jsonStr))

	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Panicf("获取xbox token: 请求失败, %s\n", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("获取xbox token: 读取Body失败, %s\n", err)
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("获取xbox token: json解析失败, %s\n", err)
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	return XBLTokenResponse{
		data["Token"].(string),
		data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string),
	}
}

func GetToken(code string) string {
	res, err := http.Post("https://login.live.com/oauth20_token.srf",
		"application/x-www-form-urlencoded",
		strings.NewReader(
			fmt.Sprintf("client_id=00000000402b5328&code=%s&grant_type=authorization_code&redirect_uri=https://login.live.com/oauth20_desktop.srf", code),
		),
	)
	if err != nil || res.StatusCode != http.StatusOK {
		log.Panicf("获取token: 请求失败, %s\n", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("获取token: 读取Body失败, %s\n", err)
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panicf("获取token: json解析失败，%s\n", err)
	}
	// log.Println(data["access_token"].(string))
	// log.Println(data["refresh_token"].(string))

	return data["access_token"].(string)
}
