package controllers

import (
	"ach/core"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


func Log(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	ws.WriteMessage(
		websocket.TextMessage,
		[]byte(core.ACH.LogBuf.GetBuf()),
	)
	core.ACH.LogWsPool.AddWs(ws)
	defer func() {
		ws.Close()
		ws = nil
	}()
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// core.ACH.ProcessInput(str)
	}
}
