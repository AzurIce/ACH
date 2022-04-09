package controllers

import (
	"ach/core"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Console(c *gin.Context) {
	// log.Println("consoleHandler")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	ws.WriteMessage(
		websocket.TextMessage,
		[]byte(strings.Join(core.ACH.OutputBuffer[core.ACH.OutputCursor:], "")+strings.Join(core.ACH.OutputBuffer[:core.ACH.OutputCursor], "")),
	)
	core.ACH.WsList = append(core.ACH.WsList, ws)
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
		core.ACH.ProcessInput(str)
	}
}
