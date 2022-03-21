package core

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ACHCore ...
type ACHCore struct {
	wg           sync.WaitGroup
	Servers      map[string]*Server
	config       *ACHConfig
	outputBuffer []string
	outputCursor int
	router       *gin.Engine
	ws           *websocket.Conn
	OutChan      chan string
}

// Ach ...
func Ach() *ACHCore {
	ach := &ACHCore{
		wg:           sync.WaitGroup{},
		Servers:      make(map[string]*Server),
		config:       DefaultACHConfig(),
		OutChan:      make(chan string, 8),
		ws:           nil,
		outputCursor: 0,
	}
	ach.init()

	return ach
}

func (ach *ACHCore) TestRun() {
	ach.router.Run(":8888")
}

// Run...
func (ach *ACHCore) Run() {
	go ach.router.Run(":8888")
	go ach.handleOut()
	ach.startServers()
	ach.wg.Wait()
}

func (ach *ACHCore) startServers() {
	for _, server := range ach.Servers {
		server.Start()
		go server.Wait()
	}
}

func (ach *ACHCore) handleOut() {
	for {
		// fmt.Println(ach.OutChan)
		str := <-ach.OutChan
		log.Print(str)
		ach.pushToBuffer(str)
		if (ach.ws != nil) {
			err := ach.ws.WriteMessage(websocket.TextMessage, []byte(str))
			if err != nil {
				// Not established.
			}
		}
	}
}

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

// --- web ---

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ach *ACHCore) handler(c *gin.Context) {
	var err error
	ach.ws, err = upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	ach.ws.WriteMessage(websocket.TextMessage, []byte(strings.Join(ach.outputBuffer, "")))
	defer func(){
		ach.ws.Close()
		ach.ws = nil
	}()
	for {
		_, str, err := ach.ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		ach.processInput(str)
	}
}

// --- init ---

func (ach *ACHCore) init() {
	ach.initRouter()
	ach.initConfig()
	ach.initServers()
	os.Mkdir(ach.config.BackupDir, 0666)
}

func (ach *ACHCore) initRouter() {
	ach.router = gin.Default()
	ach.router.Use(FrontendFileHandler())
	// ach.router.Static("/", "./assets")
	ach.router.GET("/api/console", ach.handler)
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
	ach.outputBuffer = append(ach.outputBuffer, str)
	if len(ach.outputBuffer) > 1024 {
		ach.outputBuffer = ach.outputBuffer[1:]
	}
}

func FrontendFileHandler() gin.HandlerFunc {
	fileServer := http.FileServer(gin.Dir("./assets", false))
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/custom") || strings.HasPrefix(path, "/dav") || path == "/manifest.json" {
			c.Next()
			return
		}

		// 存在的静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
