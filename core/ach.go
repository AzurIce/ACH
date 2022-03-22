package core

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	_ "ach/statik"

	"github.com/rakyll/statik/fs"
)

// ACHCore ...
type ACHCore struct {
	wg           sync.WaitGroup
	Servers      map[string]*Server
	config       *ACHConfig
	outputBuffer [1024]string
	outputCursor int
	router       *gin.Engine
	wsList       []*websocket.Conn
	OutChan      chan string
	fs           http.FileSystem
}

// Ach ...
func Ach() *ACHCore {
	ach := &ACHCore{
		wg:           sync.WaitGroup{},
		Servers:      make(map[string]*Server),
		config:       DefaultACHConfig(),
		OutChan:      make(chan string, 8),
		wsList:       make([]*websocket.Conn, 0),
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
		for index, ws := range ach.wsList {
			if ws != nil {
				sendMessage(ws, str)
			} else {
				if (index < len(ach.wsList) - 1) {
					ach.wsList = append(ach.wsList[:index], ach.wsList[index+1:]...)
				} else {
					ach.wsList = append(ach.wsList[:index])
				}
			}
		}
	}
}

func sendMessage(ws *websocket.Conn, str string) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		// Not established.
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
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	ws.WriteMessage(
		websocket.TextMessage,
		[]byte(strings.Join(ach.outputBuffer[ach.outputCursor:], "")+strings.Join(ach.outputBuffer[:ach.outputCursor], "")),
	)
	ach.wsList = append(ach.wsList, ws)
	defer func() {
		ws.Close()
		ws = nil
	}()
	for {
		_, str, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		ach.processInput(str)
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

func (ach *ACHCore) initRouter() {
	ach.router = gin.Default()
	ach.router.Use(ach.FrontendFileHandler())
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
	// ach.outputBuffer = append(ach.outputBuffer, str)
	ach.outputBuffer[ach.outputCursor] = str
	if ach.outputCursor == 1024 {
		ach.outputCursor = 0
	} else {
		ach.outputCursor++
	}
}

func (ach *ACHCore) FrontendFileHandler() gin.HandlerFunc {
	fileServer := http.FileServer(ach.fs)
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}

		// 存在的静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
