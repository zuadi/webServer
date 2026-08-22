package models

import (
	"log"
	"time"
)

// Middleware defines our custom middleware constructor signature
type Middleware func(HandlerFunc) HandlerFunc

// LoggingMiddleware is an example of a custom middleware
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx Context) {
		start := time.Now()

		// 1. Pre-execution logic (before the main handler runs)
		log.Printf("Started %s %s", ctx.request.Method, ctx.request.URL.Path)

		// 2. Pass the custom context down the chain to the next handler
		next(ctx)

		// 3. Post-execution logic
		log.Printf("Completed in %v", time.Since(start))
	}
}

func Chain(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	// Apply middlewares in reverse order so they execute in the order listed
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
