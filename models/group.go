package models

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/zuadi/webServer/constants"
	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/utils"
)

type Group struct {
	Path        string
	Route       *Route
	middlewares []Middleware
}

func (g *Group) NewGroup(path string) *Group {
	parentMWs := make([]Middleware, len(g.middlewares))
	copy(parentMWs, g.middlewares)

	return &Group{
		Path:        utils.CleanPath(g.Path) + utils.CleanPath(path),
		Route:       g.Route,
		middlewares: parentMWs,
	}
}

// AddMiddleware registers middleware specifically for this group
func (g *Group) AddMiddleware(mw ...Middleware) *Group {
	g.middlewares = append(g.middlewares, mw...)
	return g // Allows method chaining
}

func (g *Group) Get(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_GET, path)
	chainedHandler := g.chain(handler)
	g.Route.Insert(constants.METHOD_GET, path, chainedHandler)
}

func (g *Group) Post(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_POST, path)
	chainedHandler := g.chain(handler)
	g.Route.Insert(constants.METHOD_POST, path, chainedHandler)
}

func (g *Group) Put(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_PUT, path)
	chainedHandler := g.chain(handler)
	g.Route.Insert(constants.METHOD_PUT, path, chainedHandler)
}

func (g *Group) Update(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_UPDATE, path)
	chainedHandler := g.chain(handler)
	g.Route.Insert(constants.METHOD_UPDATE, path, chainedHandler)
}

func (g *Group) Delete(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_DELETE, path)
	chainedHandler := g.chain(handler)
	g.Route.Insert(constants.METHOD_DELETE, path, chainedHandler)
}

// ServeFile serves a single static file under this group's path prefix
func (g *Group) ServeFile(path, file string) {
	fullPath := utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_GET+" File", fullPath)

	g.Route.Insert(constants.METHOD_GET, fullPath, func(ctx Context) {
		http.ServeFile(ctx.GetResponseWriter(), ctx.GetRequest(), file)
	})
}

// ServeFileSystem serves a static folder under this group's path prefix
func (g *Group) ServeFileSystem(path, directory string) {
	fullPath := utils.CleanPath(g.Path) + utils.CleanPath(path)
	triePath := "/" + strings.Trim(fullPath, "/")
	stripPrefix := strings.TrimSuffix(triePath, "*")

	fs := http.FileServer(http.Dir(directory))
	handler := http.StripPrefix(stripPrefix, fs)

	logger.DebugWithStyle(constants.METHOD_GET+" "+constants.FILESYSTEM, triePath)

	g.Route.Insert(constants.METHOD_GET, triePath, func(ctx Context) {
		handler.ServeHTTP(ctx.GetResponseWriter(), ctx.GetRequest())
	})
}

func (g *Group) NewWebSocket(path string) (client *WSClient) {
	client = &WSClient{}
	fullPath := utils.CleanPath(g.Path) + utils.CleanPath(path)

	// Configure the Upgrader
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// CRITICAL: CheckOrigin allows your frontend (like localhost:3000)
		// to connect. For testing, we return true.
		CheckOrigin: func(req *http.Request) bool {
			for allowOrigin := range strings.SplitSeq(g.Route.Cors.AllowOrigins, ",") {
				allowOrigin = strings.TrimSpace(allowOrigin)

				if allowOrigin == "*" || allowOrigin == req.Header.Get("Origin") {
					return true
				}
			}
			return false
		},
	}

	handler := func(ctx Context) {
		// 1. Upgrade the HTTP connection to a WebSocket connection
		w := ctx.GetResponseWriter()
		req := ctx.GetRequest()
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			logger.ErrorWithStyle(constants.METHOD_WEBSOCKET+" "+constants.ERROR, err.Error())
			return
		}

		g.Route.Connections.Store(conn, true)

		defer func() {
			g.Route.Connections.Delete(conn)
			conn.Close()
		}()

		logger.DebugWithStyle(constants.METHOD_WEBSOCKET+" "+constants.CONNECT, req.RemoteAddr)

		// call function for example sending initial data
		client.NewConn(conn)

		// 2. The Event Loop (Keep the connection alive)
		for {
			// Read message from browser
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				logger.DebugWithStyle(constants.METHOD_WEBSOCKET+" "+constants.DISCONNECT, err.Error())
				return
			}

			// received
			if client != nil {
				client.SetConnection(conn)
				client.Recieve(p)
			}

			g.Broadcast(fullPath, messageType, p, conn)
		}
	}

	logger.DebugWithStyle(constants.METHOD_WEBSOCKET, fullPath)
	g.Route.Insert(constants.METHOD_GET, fullPath, handler)

	return
}

func (g *Group) NewBroker(path string, address string, brokerPort, wsPort int) (b *Broker, err error) {
	b = &Broker{}
	fullPath := utils.CleanPath(g.Path) + utils.CleanPath(path)

	b, err = NewBroker(address, brokerPort, wsPort)
	if err != nil {
		return
	}

	logger.DebugWithStyle(constants.METHOD_POST, path)

	g.Route.Insert(constants.METHOD_POST, fullPath, func(ctx Context) {
		r := ctx.GetRequest()
		w := ctx.GetResponseWriter()

		logger.DebugWithStyle(constants.METHOD_POST, path)
		// Read and parse JSON data from the HTTP POST request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"status":"error","message":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var msg MQTTMessage

		if err := json.Unmarshal(body, &msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if cookie, ok := r.Header["Cookie"]; ok {
			msg.ID = utils.GetCookie(cookie, "client_uuid")
		}

		msg.RemoteAddress = r.RemoteAddr

		if err := b.Publish(strings.TrimLeft(r.URL.Path, path+"/"), msg.Payload); err != nil {
			http.Error(w, `{"status":"error","message":"`+err.Error()+`"}`, http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","message":"broadcasted to mqtt loop"}`))
	})

	subPath := strings.Replace(path, "*", "", 1) + "/sub"

	logger.DebugWithStyle(constants.METHOD_POST, subPath)
	g.Route.Insert(constants.METHOD_GET, fullPath+"/sub", func(ctx Context) {
		logger.DebugWithStyle(constants.METHOD_POST, subPath)
		b.ServeWS(ctx.GetResponseWriter(), ctx.GetRequest())
	})
	return
}

func (g *Group) Broadcast(path string, messageType int, data []byte, sender *websocket.Conn) {
	g.Route.Connections.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)

		if conn == sender {
			return true
		}
		// Write the message to this specific connection
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			logger.ErrorWithStyle(constants.METHOD_WEBSOCKET+" "+constants.ERROR, "failed to send to one client")
			conn.Close()
			g.Route.Connections.Delete(conn)
		}
		return true // Continue to next connection
	})
}

// chain wraps the handler with all group middlewares in reverse order
func (g *Group) chain(handler HandlerFunc) HandlerFunc {
	return Chain(handler, g.middlewares...)
}
