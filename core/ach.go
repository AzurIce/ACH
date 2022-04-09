package core

import (
	"log"
	"os"

	"github.com/gorilla/websocket"

	"ach/bootstrap"
	"ach/models"
	_ "ach/statik"
	"ach/utils"
)

// ACHCore ...
type ACHCore struct {
	Servers           map[string]*Server
	// config            *ACHConfig
	OutputBuffer      [1024]string
	OutputCursor      int
	WsList            []*websocket.Conn
	OutChan           chan string
	// fs                http.FileSystem
	// router            *gin.Engine
	// db                *gorm.DB
}

var ACH *ACHCore

func Init() {
	ACH = &ACHCore{
		// config:            DefaultACHConfig(),
		Servers:           make(map[string]*Server),
		OutChan:           make(chan string, 8),
		WsList:            make([]*websocket.Conn, 0),
	}
	ACH.init()
}

// Ach ...
func Ach() *ACHCore {
	ach := &ACHCore{
		// config:            DefaultACHConfig(),
		Servers:           make(map[string]*Server),
		OutChan:           make(chan string, 8),
		WsList:            make([]*websocket.Conn, 0),
	}
	ach.init()

	return ach
}

func (ach *ACHCore) startAllServers() {
	for _, server := range ach.Servers {
		server.Start()
		go server.Wait()
	}
}

func (ach *ACHCore) runServer(server *Server) {
	if err := server.SStart(); err != nil {
		log.Printf("server<%s>: Error when starting:\n%s\n", server.name, err)
		return
	}

	if err := server.WWait(); err != nil {
		log.Printf("server<%s>: Error when waiting:\n%s\n", server.name, err)
	}
}

func (ach *ACHCore) handleOut() {
	for {
		// fmt.Println(ach.OutChan)
		str := <-ach.OutChan
		log.Print(str)
		ach.pushToBuffer(str)
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
	// ach.initRouter()
	// ach.initConfig()
	ach.initServers()
	models.Init()
	// ach.initDB()
	os.Mkdir(bootstrap.Config.BackupDir, 0666)
}


func (ach *ACHCore) initServers() {
	for name, serverConfig := range bootstrap.Config.Servers {
		ach.Servers[name] = NewServer(name, serverConfig, ach)
	}
}

func (ach *ACHCore) println(str string) {
	// 	log.Println(str)
	ach.OutChan <- str + "\n"
	// 	ach.pushToBuffer(str + "\n")
}

func (ach *ACHCore) pushToBuffer(str string) {
	// ach.outputBuffer = append(ach.outputBuffer, str)
	ach.OutputBuffer[ach.OutputCursor] = str
	if ach.OutputCursor == 1023 {
		ach.OutputCursor = 0
	} else {
		ach.OutputCursor++
	}
}
