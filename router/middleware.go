package router

import "github.com/zuadi/webServer/logger"

type CORSMiddleware struct {
	allowOrigins        string
	allowMethods        string
	allowHeaders        string
	allowPrivateNetwork string
}

func (r *Router) CheckCors() {
	if r.cors != nil {
		return
	}
	r.DefaultCORS()
}

func (r *Router) DefaultCORS() {
	logger.InfoWithStyle("CORS", "default CORS active please set for production")
	r.cors = &CORSMiddleware{
		allowOrigins:        "*",
		allowMethods:        "POST, GET, OPTIONS, PUT, DELETE",
		allowHeaders:        "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		allowPrivateNetwork: "true",
	}
}

func (r *Router) CORSMiddleware(cors CORSMiddleware) {
	r.cors = &cors
}
