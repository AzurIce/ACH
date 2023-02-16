package utils

import (
	"io/ioutil"
	"log"
	"strings"
)

func IsOp(opFilePath string, uuid string) bool {
	data, err := ioutil.ReadFile(opFilePath)
	if err != nil {
		log.Println("op.json读取失败")
	}

	jsonStr := string(data)
	return strings.Contains(jsonStr, uuid)
}

func IsInWhitelist(whitelistFilePath string, uuid string) bool {
	data, err := ioutil.ReadFile(whitelistFilePath)
	if err != nil {
		log.Println("whitelist.json读取失败")
	}

	jsonStr := string(data)
	return strings.Contains(jsonStr, uuid)
}
