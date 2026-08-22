package router

import (
	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/models"
)

func (r *Router) CheckCors() {
	if r.route.Cors != nil {
		return
	}
	r.DefaultCORS()
}

func (r *Router) DefaultCORS() {
	logger.InfoWithStyle("CORS", "default CORS active please set for production")
	r.route.Cors = models.DefaultCORS()
}

func (r *Router) CORSMiddleware(cors models.CORSMiddleware) {
	r.route.Cors = &cors
}
