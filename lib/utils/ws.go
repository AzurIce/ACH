package utils

import (
	"log"

	"github.com/gorilla/websocket"
)

func SendMessage(ws *websocket.Conn, str string) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		log.Println(str)
		log.Println([]byte(str))
		log.Println(err)
		// Not established.
	}
}

// WsPool
type WsPool struct {
	wsList []*websocket.Conn
}

func NewWsPool() *WsPool{
	return &WsPool{wsList: make([]*websocket.Conn, 0)}
}

func (wsPool *WsPool) AddWs(ws *websocket.Conn) {
	for index, ws := range wsPool.wsList {
		if ws == nil {
			wsPool.wsList[index] = ws
			return
		}
	}
	wsPool.wsList = append(wsPool.wsList, ws)
}

func (wsPool *WsPool) AllSendMessage(line string) {
	for _, ws := range wsPool.wsList {
		if ws != nil {
			SendMessage(ws, line)
		}
	}
}