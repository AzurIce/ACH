package core

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	_ "ach/statik"

	"github.com/rakyll/statik/fs"
)

// ACHCore ...
type ACHCore struct {
	activeServerCount sync.WaitGroup
	Servers           map[string]*Server
	config            *ACHConfig
	outputBuffer      [1024]string
	outputCursor      int
	wsList            []*websocket.Conn
	OutChan           chan string
	fs                http.FileSystem
	router            *gin.Engine
}

// TestRouter...
func (ach *ACHCore) TestRouter() {
	ach.router.Run(":8888")
}

// Ach ...
func Ach() *ACHCore {
	ach := &ACHCore{
		config:            DefaultACHConfig(),
		Servers:           make(map[string]*Server),
		activeServerCount: sync.WaitGroup{},
		OutChan:           make(chan string, 8),
		wsList:            make([]*websocket.Conn, 0),
	}
	ach.init()

	return ach
}

// Run...
func (ach *ACHCore) Run() {
	go ach.router.Run(":8888")
	go ach.handleOut()
	ach.startAllServers()
	ach.activeServerCount.Wait()
}

func (ach *ACHCore) startAllServers() {
	for _, server := range ach.Servers {
		server.Start()
		go server.Wait()
		// if !server.keepAlive {
		// 	ach.activeServerCount.Add(1)
		// }
		// go ach.runServer(server)
	}
}

func (ach *ACHCore) runServer(server *Server) {
	log.Println("added")
	if err := server.SStart(); err != nil {
		log.Printf("server<%s>: Error when starting:\n%s\n", server.name, err)
		ach.activeServerCount.Done()
		return
	}

	if err := server.WWait(); err != nil {
		log.Printf("server<%s>: Error when waiting:\n%s\n", server.name, err)
	}
	if !server.keepAlive {
		ach.activeServerCount.Done()
	}
}

func (ach *ACHCore) handleOut() {
	for {
		// fmt.Println(ach.OutChan)
		str := <-ach.OutChan
		log.Print(str)
		ach.pushToBuffer(str)
		for index, ws := range ach.wsList {
			if ws != nil {
				sendMessage(ws, str)
			} else {
				if index < len(ach.wsList)-1 {
					ach.wsList = append(ach.wsList[:index], ach.wsList[index+1:]...)
				} else {
					ach.wsList = append(ach.wsList[:index])
				}
			}
		}
	}
}

//
func (ach *ACHCore) processInput(line []byte) {
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
	ach.initStaticFS()
	ach.initRouter()
	ach.initConfig()
	ach.initServers()
	os.Mkdir(ach.config.BackupDir, 0666)
}

func (ach *ACHCore) initStaticFS() {
	var err error
	ach.fs, err = fs.New()
	if err != nil {
		log.Panicln(err)
	}
}

func (ach *ACHCore) initConfig() {
	err := ach.readConfig()
	if err != nil {
		if os.IsNotExist(err) { // 文件不存在，创建并写入默认配置
			ach.println("[ACH]: Cannot find config.yml, creating...")
			ach.saveConfig()
			ach.println("[ACH]: Successful created config.yml, please complete the config.")
		}
		os.Exit(1)
	}
}

func (ach *ACHCore) initServers() {
	for name, serverConfig := range ach.config.Servers {
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
	ach.outputBuffer[ach.outputCursor] = str
	if ach.outputCursor == 1024 {
		ach.outputCursor = 0
	} else {
		ach.outputCursor++
	}
}
