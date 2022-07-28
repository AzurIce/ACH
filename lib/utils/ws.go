package utils

import (
	"log"

	"github.com/gorilla/websocket"
)

func SendMessage(ws *websocket.Conn, str string) error {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		log.Println("ERR: ", str)
		log.Println("ERR: ", []byte(str))
		log.Println("ERR: ", err)
		ws.Close()
		// Not established.
	}
	return err
}

// WsPool
type WsPool struct {
	wsList []*websocket.Conn
}

func NewWsPool() *WsPool{
	return &WsPool{wsList: make([]*websocket.Conn, 0)}
}

func (wsPool *WsPool) AddWs(newWs *websocket.Conn) {
	for i, ws := range wsPool.wsList {
		if ws == nil {
			wsPool.wsList[i] = newWs
			return
		}
	}
	wsPool.wsList = append(wsPool.wsList, newWs)
}

func (wsPool *WsPool) AllSendMessage(line string) {
	// log.Println(wsPool.wsList)
	for i, ws := range wsPool.wsList {
		if ws != nil {
			if err := SendMessage(ws, line); err != nil {
				wsPool.wsList[i] = nil
			}
		}
	}
}