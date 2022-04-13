package utils

import (
	"log"

	"github.com/gorilla/websocket"
)

func SendMessage(ws *websocket.Conn, str string) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		log.Println(err)
		// Not established.
	}
}