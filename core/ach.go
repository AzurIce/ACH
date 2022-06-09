package core

import (
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"

	"ach/bootstrap"
	"ach/lib/utils"
	"ach/models"
	_ "ach/statik"
)

// ACHCore ...
type ACHCore struct {
	Servers      map[string]*Server
	LogBuf       [1024]string
	LogBufCursor int
	LogWsList    []*websocket.Conn
	OutputBuffer [1024]string
	OutputCursor int
	WsList       []*websocket.Conn
	OutChan      chan string
	LogChan      chan string
}

var ACH *ACHCore

func Init() {
	ACH = &ACHCore{
		Servers:   make(map[string]*Server),
		OutChan:   make(chan string, 8),
		LogChan:   make(chan string, 8),
		LogWsList: make([]*websocket.Conn, 0),
		WsList:    make([]*websocket.Conn, 0),
	}
	ACH.init()
}

func (ach *ACHCore) StartAllServers() {
	log.Println("[ACHCore]: Starting all servers")
	go ach.handleOut()
	for _, server := range ach.Servers {
		go server.Run()
		// go ach.runServer(server)
	}
}

// func (ach *ACHCore) runServer(server *Server) {
// 	if err := server.Start(); err != nil {
// 		log.Printf("server<%s>: Error when starting:\n%s\n", server.name, err)
// 		return
// 	}

// 	if err := server.Wait(); err != nil {
// 		log.Printf("server<%s>: Error when waiting:\n%s\n", server.name, err)
// 	}
// }

func (ach *ACHCore) handleOut() {
	for {
		select {
		case str := <-ach.OutChan:
			log.Print(str)
			for index, ws := range ach.WsList {
				if ws != nil {
					utils.SendMessage(ws, str)
				} else {
					if index < len(ach.WsList)-1 {
						ach.WsList = append(ach.WsList[:index], ach.WsList[index+1:]...)
					} else {
						ach.WsList = ach.WsList[:index]
					}
				}
			}
		case str := <-ach.LogChan:
			log.Print(str)
			for index, ws := range ach.LogWsList {
				if ws != nil {
					utils.SendMessage(ws, str)
				} else {
					if index < len(ach.WsList)-1 {
						ach.WsList = append(ach.WsList[:index], ach.WsList[index+1:]...)
					} else {
						ach.WsList = ach.WsList[:index]
					}
				}
			}
		}
	}
}

//
func (ach *ACHCore) ProcessInput(line []byte) {
	// 转发正则
	res := ForwardReg.FindSubmatch(line)
	if res != nil { // 转发到特定服务器
		server, exist := ach.Servers[string(res[1])]
		if exist {
			server.InChan <- string(res[2])
		} else {
			log.Printf("MCSH[stdinForward/ERROR]: Cannot find running server <%v>\n", string(res[1]))
		}
	} else { // 转发到所有服务器
		for _, server := range ach.Servers {
			server.InChan <- string(line)
		}
	}
}

// --- init ---

func (ach *ACHCore) init() {
	bootstrap.InitStaticFS()
	bootstrap.InitConfig()
	ach.initServers()
	models.Init()
	os.Mkdir(bootstrap.Config.BackupDir, 0666)
}

func (ach *ACHCore) initServers() {
	for name, serverConfig := range bootstrap.Config.Servers {
		ach.Servers[name] = NewServer(name, serverConfig)
	}
}

func (ach *ACHCore) Log(str string) {
	ach.LogBuf[ach.LogBufCursor] = str
	ach.LogBufCursor = (ach.LogBufCursor + 1) % 1024
	ach.LogChan <- str
}

func (ach *ACHCore) Println(str string) {
	ach.OutputBuffer[ach.OutputCursor] = str
	ach.OutputCursor = (ach.OutputCursor + 1) % 1024
	ach.OutChan <- str
}

func (ach *ACHCore) GetLog() string {
	return strings.Join(ach.LogBuf[ach.LogBufCursor:], "") + strings.Join(ach.LogBuf[:ACH.LogBufCursor], "")
}
