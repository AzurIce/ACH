package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
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

func GetPlayerInfoByCode(code string) (PlayerInfo, error) {
	var playerInfo PlayerInfo

	// 获取账号对应 UUID
	// 1. Code -> Token
	token, err := GetToken(code)
	if err != nil {
		return playerInfo, err
	}
	// 2. Token -> XBL Token
	xblRes, err := GetXBLToken(token)
	if err != nil {
		return playerInfo, err
	}
	// 3. XBL Token -> XSTS Token
	xstsRes, err := GetXSTSToken(xblRes.Token)
	if err != nil {
		return playerInfo, err
	}
	// 4. XSTS Token + UHS -> MC Token
	mcToken, err := GetMCToken(xstsRes)
	if err != nil {
		return playerInfo, err
	}
	// 5. MC Token -> Player Info
	playerInfo, err = GetPlayerInfo(mcToken)
	if err != nil {
		return playerInfo, err
	}

	return playerInfo, nil
}

func GetPlayerInfo(mcToken string) (PlayerInfo, error) {
	var playerInfo PlayerInfo
	log.Println("获取player info...")
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
		log.Printf("获取player info: 请求失败, %s\n", err)
		return playerInfo, errors.New("err")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("获取player info: 读取Body失败, %s\n", err)
		return playerInfo, errors.New("err")
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("获取player info: json解析失败, %s\n", err)
		return playerInfo, errors.New("err")
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	log.Println("获取完成")
	playerInfo = PlayerInfo{
		data["id"].(string),
		data["name"].(string),
	}
	return playerInfo, nil
}

func GetMCToken(xsts XSTSTokenResponse) (string, error) {
	var mcToken string
	log.Println("获取mc token...")

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
		log.Printf("获取mc token: 请求失败, %s\n", err)
		return mcToken, errors.New("err")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("获取mc token: 读取Body失败, %s\n", err)
		return mcToken, errors.New("err")
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("获取mc token: json解析失败, %s\n", err)
		return mcToken, errors.New("err")
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	log.Println("获取完成")
	mcToken = data["access_token"].(string)
	return mcToken, nil
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

func GetXSTSToken(xblToken string) (XSTSTokenResponse, error) {
	var xstsTokenResponse XSTSTokenResponse
	log.Println("获取xsts token...")
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
		log.Printf("获取xsts token: 请求失败, %s\n", err)
		return xstsTokenResponse, errors.New("err")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("获取xsts token: 读取Body失败, %s\n", err)
		return xstsTokenResponse, errors.New("err")
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("获取xsts token: json解析失败, %s\n", err)
		return xstsTokenResponse, errors.New("err")
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	log.Println("获取完成")
	xstsTokenResponse = XSTSTokenResponse{
		data["Token"].(string),
		data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string),
	}
	return xstsTokenResponse, nil
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

func GetXBLToken(token string) (XBLTokenResponse, error) {
	var xblTokenResponse XBLTokenResponse

	log.Println("获取xbl token...")
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
		log.Printf("获取xbox token: 请求失败, %s\n", err)
		return xblTokenResponse, errors.New("err")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("获取xbox token: 读取Body失败, %s\n", err)
		return xblTokenResponse, errors.New("err")
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("获取xbox token: json解析失败, %s\n", err)
		return xblTokenResponse, errors.New("err")
	}
	// log.Println(data["Token"].(string))
	// log.Println(data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string))

	log.Println("获取完成")
	xblTokenResponse = XBLTokenResponse{
		data["Token"].(string),
		data["DisplayClaims"].(map[string]interface{})["xui"].([]interface{})[0].(map[string]interface{})["uhs"].(string),
	}
	return xblTokenResponse, nil
}

func GetToken(code string) (string, error) {
	var token string

	log.Println("获取token...")
	res, err := http.Post("https://login.live.com/oauth20_token.srf",
		"application/x-www-form-urlencoded",
		strings.NewReader(
			fmt.Sprintf("client_id=00000000402b5328&code=%s&grant_type=authorization_code&redirect_uri=https://login.live.com/oauth20_desktop.srf", code),
		),
	)
	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("获取token: 请求失败, %s\n", err)
		return token, errors.New("err")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("获取token: 读取Body失败, %s\n", err)
		return token, errors.New("err")
	}
	// log.Println(string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("获取token: json解析失败, %s\n", err)
		return token, errors.New("err")
	}
	token = data["access_token"].(string)
	// log.Println(data["access_token"].(string))
	// log.Println(data["refresh_token"].(string))

	log.Println("获取完成")
	return token, nil
}
