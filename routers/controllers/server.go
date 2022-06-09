package controllers

import (
	"ach/core"
	"log"
	"net/http"

	"ach/services/server"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Console(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// log.Println("consoleHandler")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("[Console]: upgrade:", err)
		return
	}
	ws.WriteMessage(
		websocket.TextMessage,
		[]byte(core.ACH.OutBuf.GetBuf()),
	)
	core.ACH.OutWsPool.AddWs(ws)
	defer func() {
		ws.Close()
		ws = nil
	}()
	for {
		_, str, err := ws.ReadMessage()
		if err != nil {
			log.Println("[Console]: read:", err)
			break
		}
		core.ACH.InChan <- string(str)
		// core.ACH.ProcessInput(str)
	}
}

func GetServers(c *gin.Context) {
	log.Println(server.GetServers())
	c.JSON(http.StatusOK, server.GetServers())
}
