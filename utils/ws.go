package utils

import "github.com/gorilla/websocket"

func SendMessage(ws *websocket.Conn, str string) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		// Not established.
	}
}