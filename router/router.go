package router

import (
	"net/http"
	"strings"

	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/models"
	"github.com/zuadi/webServer/utils"
)

type Router struct {
	route models.Route
	cors  *CORSMiddleware
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
