package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/zuadi/webServer/constants"
	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/models"
	"github.com/zuadi/webServer/utils"
)

type Router struct {
	route models.Route
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ServeFile(path, file string) {
	logger.DebugWithStyle(constants.METHOD_GET+" File", path)

	r.route.Insert(constants.METHOD_GET, path, func(ctx models.Context) {
		logger.DebugWithStyle(constants.METHOD_GET+" File", path)
		http.ServeFile(ctx.GetResponseWriter(), ctx.GetRequest(), file)
	})
}

/*
serves file system specified in directory

	directory = hst to start with . if in current directory
	path = server url, if all underlying file and folder should be browsable path/* * hast to be added
*/
func (r *Router) ServeFileSystem(path, directory string) {
	triePath := "/" + strings.Trim(path, "/")
	stripPrefix := strings.TrimSuffix(triePath, "*")

	fs := http.FileServer(http.Dir(directory))
	handler := http.StripPrefix(stripPrefix, fs)

	logger.DebugWithStyle(constants.METHOD_GET+" "+constants.FILESYSTEM, triePath)

	r.route.Insert(constants.METHOD_GET, triePath, func(ctx models.Context) {
		logger.DebugWithStyle(constants.METHOD_GET+" "+constants.FILESYSTEM, ctx.GetRequest().URL.Path)
		handler.ServeHTTP(ctx.GetResponseWriter(), ctx.GetRequest())
	})
}

func (r *Router) NewGroup(path string) *models.Group {
	return &models.Group{
		Path:  utils.CleanPath(path),
		Route: &r.route,
	}
}

func (r *Router) Get(path string, handler models.HandlerFunc) {
	logger.DebugWithStyle(constants.METHOD_GET, path)
	r.route.Insert(constants.METHOD_GET, utils.CleanPath(path), handler)
}

func (r *Router) Post(path string, handler models.HandlerFunc) {
	logger.DebugWithStyle(constants.METHOD_POST, path)
	r.route.Insert(constants.METHOD_POST, utils.CleanPath(path), handler)
}

func (r *Router) Put(path string, handler models.HandlerFunc) {
	logger.DebugWithStyle(constants.METHOD_PUT, path)
	r.route.Insert(constants.METHOD_PUT, utils.CleanPath(path), handler)
}

func (r *Router) Update(path string, handler models.HandlerFunc) {
	logger.DebugWithStyle(constants.METHOD_UPDATE, path)
	r.route.Insert(constants.METHOD_UPDATE, utils.CleanPath(path), handler)
}

func (r *Router) Delete(path string, handler models.HandlerFunc) {
	logger.DebugWithStyle(constants.METHOD_DELETE, path)
	r.route.Insert(constants.METHOD_DELETE, utils.CleanPath(path), handler)
}

func (r *Router) NewWebSocket(path string) (client *models.WSClient) {
	client = &models.WSClient{}
	cleanPath := utils.CleanPath(path)

	// Configure the Upgrader
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// CRITICAL: CheckOrigin allows your frontend (like localhost:3000)
		// to connect. For testing, we return true.
		CheckOrigin: func(req *http.Request) bool {
			for allowOrigin := range strings.SplitSeq(r.route.Cors.AllowOrigins, ",") {
				allowOrigin = strings.TrimSpace(allowOrigin)

				if allowOrigin == "*" || allowOrigin == req.Header.Get("Origin") {
					return true
				}
			}
			return false
		},
	}

	handler := func(ctx models.Context) {
		// 1. Upgrade the HTTP connection to a WebSocket connection
		w := ctx.GetResponseWriter()
		req := ctx.GetRequest()
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			logger.ErrorWithStyle(constants.METHOD_WEBSOCKET+" "+constants.ERROR, err.Error())
			return
		}

		r.route.Connections.Store(conn, true)

		defer func() {
			r.route.Connections.Delete(conn)
			conn.Close()
		}()

		logger.DebugWithStyle(constants.METHOD_WEBSOCKET+" "+constants.CONNECT, req.RemoteAddr)

		// call function for example sending initial data
		client.NewConn(conn)

		// 2. Send broadcast message
		client.Send = func(messageType models.MessageType, data []byte) {
			r.Broadcast(cleanPath, int(messageType), data, nil)
		}

		// 3. The Event Loop (Keep the connection alive)
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

			r.Broadcast(cleanPath, messageType, p, conn)
		}
	}

	logger.DebugWithStyle(constants.METHOD_WEBSOCKET, path)
	r.route.Insert(constants.METHOD_GET, cleanPath, handler)

	return
}

// NewBroker is initiating a new mqtt broker
func (r *Router) NewBroker(path string, address string, brokerPort, wsPort int) (b *models.Broker, err error) {
	b = &models.Broker{}
	cleanPath := utils.CleanPath(path)
	fmt.Println(cleanPath)

	b, err = models.NewBroker(address, brokerPort, wsPort)
	if err != nil {
		return
	}

	logger.DebugWithStyle(constants.METHOD_POST, path)

	r.route.Insert(constants.METHOD_POST, cleanPath, func(ctx models.Context) {
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

		var msg models.MQTTMessage

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
	r.route.Insert(constants.METHOD_GET, cleanPath+"/sub", func(ctx models.Context) {
		logger.DebugWithStyle(constants.METHOD_POST, subPath)
		b.ServeWS(ctx.GetResponseWriter(), ctx.GetRequest())
	})
	return
}

func (r *Router) Broadcast(path string, messageType int, data []byte, sender *websocket.Conn) {
	r.route.Connections.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)

		if conn == sender {
			return true
		}
		// Write the message to this specific connection
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			logger.ErrorWithStyle(constants.METHOD_WEBSOCKET+" "+constants.ERROR, "failed to send to one client")
			conn.Close()
			r.route.Connections.Delete(conn)
		}
		return true // Continue to next connection
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")

	for allowOrigin := range strings.SplitSeq(r.route.Cors.AllowOrigins, ",") {
		allowOrigin = strings.TrimSpace(allowOrigin)

		if allowOrigin == "*" || allowOrigin == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			break
		}
	}

	w.Header().Set("Access-Control-Allow-Methods", r.route.Cors.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", r.route.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Allow-Private-Network", r.route.Cors.AllowPrivateNetwork)

	if req.Method == constants.OPTIONS {
		logger.DebugWithStyle(constants.OPTIONS, req.URL.Path)
		w.WriteHeader(http.StatusOK)
		return
	}

	found, handler, params := r.route.Search(req.Method, req.URL.Path)

	if !found || handler == nil {
		http.NotFound(w, req)
		return
	}
	var ctx models.Context
	ctx.SetRequest(req)
	ctx.SetResponseWriter(w)
	ctx.SetParameters(params)
	logger.DebugWithStyle(req.Method, req.URL.Path)
	handler(ctx)

}
