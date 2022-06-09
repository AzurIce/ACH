package server

import (
	"ach/models"
	"encoding/json"
	"log"
)

func GetServers() string {
	servers := models.GetServers()
	log.Println("[services/server/GetServers]: ", servers)
	bytes, _ := json.Marshal(servers)
	return string(bytes)
}