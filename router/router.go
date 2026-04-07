package router

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/models"
	"github.com/zuadi/webServer/utils"
)

type Router struct {
	route       models.Route
	cors        *CORSMiddleware
	connections sync.Map
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ServeFile(path, file string) {
	r.route.Insert("GET", path, func(ctx models.Context) {
		http.ServeFile(ctx.GetResponseWriter(), ctx.GetRequest(), file)
	})
}

func (r *Router) ServeFileSystem(path, directory string) {
	triePath := "/" + strings.Trim(path, "/")
	stripPrefix := strings.TrimSuffix(triePath, "*")

	fs := http.FileServer(http.Dir(directory))
	handler := http.StripPrefix(stripPrefix, fs)

	r.route.Insert("GET", triePath, func(ctx models.Context) {
		handler.ServeHTTP(ctx.GetResponseWriter(), ctx.GetRequest())
	})
}

func (r *Router) Group(path string) *models.Group {
	return &models.Group{
		Path:  utils.CleanPath(path),
		Route: &r.route,
	}
}

func (r *Router) Get(path string, handler models.Handler) {
	title := "GET"
	logger.SetStyle(title, "#56a7f8", path)
	r.route.Insert(title, utils.CleanPath(path), handler)
}

func (r *Router) Post(path string, handler models.Handler) {
	title := "POST"
	logger.SetStyle(title, "#56f8ba", path)
	r.route.Insert(title, utils.CleanPath(path), handler)
}

func (r *Router) WebSocket(path string, recieve func(data any)) {
	cleanPath := utils.CleanPath(path)

	// Configure the Upgrader
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// CRITICAL: CheckOrigin allows your frontend (like localhost:3000)
		// to connect. For testing, we return true.
		CheckOrigin: func(req *http.Request) bool {
			for allowOrigin := range strings.SplitSeq(r.cors.allowOrigins, ",") {
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
			logger.SetStyle("WS ERROR", "#ff0000", err.Error())
			return
		}

		r.connections.Store(conn, true)

		defer func() {
			r.connections.Delete(conn)
			conn.Close()
		}()

		logger.SetStyle("WS CONNECT", "#ff8800", req.RemoteAddr)

		// 2. The Event Loop (Keep the connection alive)
		for {
			// Read message from browser
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				logger.SetStyle("WS DISCONNECT", "#ff8800", err.Error())
				return
			}

			// received
			if recieve != nil {
				recieve(p)
			}

			r.Broadcast(cleanPath, messageType, p)
		}
	}

	logger.SetStyle("WS", "#ff8800", path)
	r.route.Insert("GET", cleanPath, handler)
}

func (r *Router) Broadcast(path string, messageType int, data []byte) {
	r.connections.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)

		// Write the message to this specific connection
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			logger.SetStyle("WS ERR", "#ff0000", "Failed to send to one client")
			conn.Close()
			r.connections.Delete(conn)
		}
		return true // Continue to next connection
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")

	for allowOrigin := range strings.SplitSeq(r.cors.allowOrigins, ",") {
		allowOrigin = strings.TrimSpace(allowOrigin)

		if allowOrigin == "*" || allowOrigin == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			break
		}
	}

	w.Header().Set("Access-Control-Allow-Methods", r.cors.allowMethods)
	w.Header().Set("Access-Control-Allow-Headers", r.cors.allowHeaders)
	w.Header().Set("Access-Control-Allow-Private-Network", r.cors.allowPrivateNetwork)

	if req.Method == "OPTIONS" {
		logger.SetStyle("OPTIONS", "#43ecf8", req.URL.Path)
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
	handler(ctx)

}
