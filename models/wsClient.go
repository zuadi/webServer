package models

import (
	"github.com/gorilla/websocket"
	"github.com/zuadi/webServer/constants"
	"github.com/zuadi/webServer/logger"
)

type WSClient struct {
	recieve       func(data any)
	conn          *websocket.Conn
	NewConnection func(ws *WSClient)
}

func (ws *WSClient) SetConnection(conn *websocket.Conn) {
	ws.conn = conn
}

func (ws *WSClient) Recieve(data any) {
	if ws.recieve != nil {
		ws.recieve(data)
	}
}

func (ws *WSClient) Listen(cb func(data any)) {
	ws.recieve = cb
}

func (ws *WSClient) Answer(messageType int, data []byte) {
	if ws.conn == nil {
		return
	}
	err := ws.conn.WriteMessage(messageType, data)
	if err != nil {
		logger.ErrorWithStyle(constants.METHOD_WEBSOCKET+" "+constants.ERROR, "failed to send to one client")
		ws.conn.Close()
	}
}

func (ws *WSClient) NewConn(conn *websocket.Conn) {
	if conn == nil || ws.NewConnection == nil {
		return
	}
	ws.NewConnection(&WSClient{conn: conn})
}
