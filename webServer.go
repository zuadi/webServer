package webServer

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zuadi/webServer.git/models"
	"github.com/zuadi/webServer.git/router"
)

type WebServer struct {
	url    string
	router *router.Router
}

func NewWebServer(ip string, port int) *WebServer {
	return &WebServer{
		url:    fmt.Sprintf("%s:%d", ip, port),
		router: router.NewRouter(),
	}
}

func (s *WebServer) ServeFile(path, file string) {
	s.router.ServeFile(path, file)
}

func (s *WebServer) ServeFileSystem(path, file string) {
	fmt.Println(12)
	s.router.ServeFileSystem(path, file)
}

func (s *WebServer) Group(name string) *models.Group {
	return s.router.Group(name)
}

func (s *WebServer) Get(name string, handler models.Handler) {
	s.router.Get(name, handler)
}

func (s *WebServer) Post(name string, handler models.Handler) {
	s.router.Post(name, handler)
}

func (s *WebServer) ListenHttp() error {
	log.Printf("listens on: %s\n", s.url)
	return http.ListenAndServe(s.url, s.router)
}
