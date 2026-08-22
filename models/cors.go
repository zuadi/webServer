package models

type CORSMiddleware struct {
	AllowOrigins        string
	AllowMethods        string
	AllowHeaders        string
	AllowPrivateNetwork string
}

func DefaultCORS() *CORSMiddleware {
	return &CORSMiddleware{
		AllowOrigins:        "*",
		AllowMethods:        "POST, GET, OPTIONS, PUT, DELETE",
		AllowHeaders:        "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		AllowPrivateNetwork: "true",
	}
}
