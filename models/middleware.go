package models

import (
	"log"
	"time"
)

// Middleware defines our custom middleware constructor signature
type Middleware func(next HandlerFunc) HandlerFunc

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

// BearerAuthMiddleware validates incoming Authorization: Bearer <token> headers.
// It accepts a validator function so you can inject custom JWT or database token checks.
func BearerAuthMiddleware(validateToken func(token string) bool) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) {
			req := ctx.GetRequest()
			w := ctx.GetResponseWriter()

			// 1. Fetch Authorization header
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
				return // Abort middleware chain
			}

			// 2. Verify "Bearer " prefix and trim it
			token, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found {
				http.Error(w, `{"error":"authorization header must start with Bearer"}`, http.StatusUnauthorized)
				return // Abort middleware chain
			}

			token = strings.TrimSpace(token)
			if token == "" {
				http.Error(w, `{"error":"bearer token cannot be empty"}`, http.StatusUnauthorized)
				return // Abort middleware chain
			}

			// 3. Validate token against business logic
			if !validateToken(token) {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return // Abort middleware chain
			}

			// 4. Token valid -> Proceed down the chain to the endpoint handler
			next(ctx)
		}
	}
}

func Chain(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	// Apply middlewares in reverse order so they execute in the order listed
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
