package models

import "ach/core"

type Server struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
}

func GetServers() []Server {
	servers := make([]Server, 0, len(core.ACH.Servers))
	for _, server := range core.ACH.Servers {
		servers = append(servers, Server{Name: server.Name, Running: server.Running})
	}
	return servers
}
