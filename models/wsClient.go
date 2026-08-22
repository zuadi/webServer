package models

import (
	"github.com/gorilla/websocket"
	"github.com/zuadi/webServer/constants"
	"github.com/zuadi/webServer/logger"
)

type MessageType int

// The message types are defined in RFC 6455.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage MessageType = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage MessageType = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage MessageType = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage MessageType = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage MessageType = 10
)

type WSClient struct {
	recieve       func(data any)
	Send    func(messageType MessageType, data []byte)
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

func (ws *WSClient) Broadcast(messageType MessageType, data []byte) {
	if ws.Send != nil {
		ws.Send(messageType, data)
	}
}

func (ws *WSClient) Listen(cb func(data any)) {
	ws.recieve = cb
}

func (ws *WSClient) Answer(messageType MessageType, data []byte) {
	if ws.conn == nil {
		return
	}
	err := ws.conn.WriteMessage(int(messageType), data)
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
