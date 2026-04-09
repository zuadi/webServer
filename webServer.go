package webServer

import (
	"fmt"
	"log"
	"net"
	"net/http"

	logger "github.com/charmbracelet/log"
	logging "github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/models"
	"github.com/zuadi/webServer/router"
)

type WebServer struct {
	ip     string
	port   int
	router *router.Router
}

func NewWebServer(ip string, port int) *WebServer {
	return &WebServer{
		ip:     ip,
		port:   port,
		router: router.NewRouter(),
	}
}

func (s *WebServer) SetLogLevel(level logger.Level) {
	logger.SetLevel(level)
	logging.DebugWithStyle("NEW ROUTER", "Debug Mode active")
}

func (s *WebServer) SetDefaultCORS() {
	s.router.DefaultCORS()
}

func (s *WebServer) ServeFile(path, file string) {
	s.router.ServeFile(path, file)
}

func (s *WebServer) ServeFileSystem(path, file string) {
	s.router.ServeFileSystem(path, file)
}

func (s *WebServer) Group(path string) *models.Group {
	return s.router.Group(path)
}

func (s *WebServer) Get(path string, handler models.Handler) {
	s.router.Get(path, handler)
}

func (s *WebServer) Post(path string, handler models.Handler) {
	s.router.Post(path, handler)
}

func (s *WebServer) WebSocket(path string, reviece func(data any)) {
	s.router.WebSocket(path, reviece)
}

func (s *WebServer) ListenHttp() error {
	s.router.CheckCors()
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return err
	}
	// Extract the actual port
	addr := ln.Addr().(*net.TCPAddr)

	log.Printf("listens on: %s:%d\n", addr.IP.String(), addr.Port)
	return http.Serve(ln, s.router)
}
