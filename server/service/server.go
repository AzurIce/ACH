package service

import (
	"ach/core"
	"ach/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GetServersService struct{}

func (s *GetServersService) Handle(c *gin.Context) (any, error) {
	servers := models.GetServers()
	log.Println("[services/server/GetServers]: ", servers)
	// bytes, _ := json.Marshal(servers)
	return servers, nil
}

type ServerConsoleService struct{}

func (s *ServerConsoleService) Handle(c *gin.Context) (any, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// log.Println("consoleHandler")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("[Console]: upgrade:", err)
		return nil, err
	}
	err = ws.WriteMessage(
		websocket.TextMessage,
		[]byte(core.ACH.OutBuf.GetBuf()),
	)
	if err != nil {
		log.Println("[Console]: write:", err)
		return nil, err
	}
	core.ACH.OutWsPool.AddWs(ws)
	defer func() {
		ws.Close()
		ws = nil
	}()
	for {
		_, str, err := ws.ReadMessage()
		if err != nil {
			log.Println("[Console]: read:", err)
			ws.Close()
			break
		}
		core.ACH.InChan <- string(str)
		// core.ACH.ProcessInput(str)
	}
	return nil, nil
}
