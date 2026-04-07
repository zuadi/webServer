package router

type CORSMiddleware struct {
	allowOrigins        string
	allowMethods        string
	allowHeaders        string
	allowPrivateNetwork string
}

func (r *Router) DefaultCORS() {
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
