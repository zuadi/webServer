package router

import (
	"net/http"
	"strings"
	"webServer/models"
)

type Router struct {
	route models.Route
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
		Name:  cleanPath(path),
		Route: &r.route,
	}
}

func (r *Router) Get(path string, handler models.Handler) {
	r.route.Insert("GET", cleanPath(path), handler)
}

func (r *Router) Post(path string, handler models.Handler) {
	r.route.Insert("POST", cleanPath(path), handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

func cleanPath(p string) string {
	if p == "/" {
		return "/"
	}
	return "/" + strings.Trim(p, "/")
}
