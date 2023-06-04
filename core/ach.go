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

	OutChan   chan string
	OutBuf    *utils.ScrollBuffer

	SSEChanList []*chan string

	InChan chan string
}

var ACH *ACHCore

func Init() {
	ACH = &ACHCore{
		Servers: make(map[string]*Server),
		InChan:  make(chan string, 8),

		OutChan:   make(chan string, 8),
		OutBuf:    utils.NewScrollBuffer(),

		LogChan:   make(chan string, 8),
		LogBuf:    utils.NewScrollBuffer(),
	}

	// for name, serverConfig := range bootstrap.Config.Servers {
	// 	ACH.Servers[name] = NewServer(name, serverConfig)
	// }

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
				if server, exist := ach.Servers[string(res[1])]; exist { // 对应服务器正在运行，则直接转发
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
			for _, c := range ach.SSEChanList {
				if c != nil {
					*c <- line
				}
			}
			// ach.OutWsPool.AllSendMessage(line)
		case line := <-ach.LogChan:
			log.Print("[ACHCore/tick]: LogChan: ", line)
			for _, c := range ach.SSEChanList {
				if c != nil {
					*c <- line
				}
			}
			// ach.LogWsPool.AllSendMessage(line)
		}
	}
}

func (ach *ACHCore) AddSSEChan(ch *chan string) {
	for i, ws := range ach.SSEChanList {
		if ws == nil {
			ach.SSEChanList[i] = ch
			return
		}
	}
	ach.SSEChanList = append(ach.SSEChanList, ch)
}

func (ach *ACHCore) RemoveSSEChan(ch *chan string) {
    for i, channel := range ach.SSEChanList {
        if channel == ch {
            ach.SSEChanList[i] = nil
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
