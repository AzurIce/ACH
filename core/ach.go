package core

import (
	"fmt"
	"log"
	"regexp"

	"ach/internal/bootstrap"
	"ach/internal/utils"
	_ "ach/statik"
)

var ForwardReg = regexp.MustCompile(`(.+?) *\| *(.+)`)

// ACHCore ...
type ACHCore struct {
	Servers map[string]*Server

	LogChan   chan string
	LogBuf    *utils.ScrollBuffer
	LogWsPool *utils.WsPool

	OutChan   chan string
	OutBuf    *utils.ScrollBuffer
	OutWsPool *utils.WsPool

	InChan chan string
}

var ACH *ACHCore

func Init() {
	ACH = &ACHCore{
		Servers: make(map[string]*Server),
		InChan:  make(chan string, 8),

		OutChan:   make(chan string, 8),
		OutBuf:    utils.NewScrollBuffer(),
		OutWsPool: utils.NewWsPool(),

		LogChan:   make(chan string, 8),
		LogBuf:    utils.NewScrollBuffer(),
		LogWsPool: utils.NewWsPool(),
	}

	for name, serverConfig := range bootstrap.Config.Servers {
		ACH.Servers[name] = NewServer(name, serverConfig)
	}

	go ACH.tick()
}

func (ach *ACHCore) StartAllServers() {
	log.Println("[ACHCore]: Starting all servers")
	for _, server := range ach.Servers {
		go server.Run()
	}
}

func (ach *ACHCore) tick() {
	for {
		select {
		case line := <-ach.InChan:
			// 转发正则
			res := ForwardReg.FindStringSubmatch(line)
			if res != nil { // 转发到特定服务器
				server, exist := ach.Servers[string(res[1])]
				if exist {
					server.InChan <- string(res[2])
				} else {
					log.Printf("[ACHCore/tick]: Cannot find running server <%v>\n", string(res[1]))
				}
			} else { // 转发到所有服务器
				for _, server := range ach.Servers {
					server.InChan <- string(line)
				}
			}
		case line := <-ach.OutChan:
			fmt.Print("[ACHCore/tick]: OutChan: ", line)
			ach.OutWsPool.AllSendMessage(line)
		case line := <-ach.LogChan:
			log.Print("[ACHCore/tick]: LogChan: ", line)
			ach.LogWsPool.AllSendMessage(line)
		}
	}
}

// --- init ---

func (ach *ACHCore) Log(str string) {
	ach.LogBuf.Write(str)
	ach.LogChan <- str
}

func (ach *ACHCore) Logln(str string) {
	ach.Log(str + "\n")
}

func (ach *ACHCore) Write(str string) {
	ach.OutBuf.Write(str)
	ach.OutChan <- str
}

func (ach *ACHCore) Writeln(str string) {
	ach.Write(str + "\n")
}
