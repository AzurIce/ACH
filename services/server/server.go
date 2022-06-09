package server

import (
	"ach/models"
	"encoding/json"
	"log"
)

func GetServers() string {
	servers := models.GetServers()
	log.Println(servers)
	bytes, _ := json.Marshal(servers)
	return string(bytes)
}