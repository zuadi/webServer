package test

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/zuadi/webServer"
	"github.com/zuadi/webServer/models"
)

// Helper function to setup a baseline server for testing basic routing
func setupBaseServer(port int) *webServer.WebServer {
	ws := webServer.NewWebServer("127.0.0.1", port)
	ws.SetLogLevel(log.DebugLevel)

	// GET & POST handlers
	ws.Get("/test1", func(ctx models.Context) { ctx.RespondString("hello from test1") })
	ws.Post("/test1", func(ctx models.Context) { ctx.RespondString("hello from test1 post") })

	// JSON response handler
	ws.Get("/test2", func(ctx models.Context) {
		var data struct {
			Info    string
			Message string
		}
		data.Info = "OK"
		data.Message = "This is a message"
		ctx.RespondJson(http.StatusOK, data)
	})

	// Routing Groups
	g := ws.Group("v2")
	g.Get("hallo", func(ctx models.Context) { ctx.RespondString("H") })
	g.Get("velo", func(ctx models.Context) { ctx.RespondString("hallo velo") })

	g2 := g.Group("v1")
	g2.Get("/:id/23/hallo", func(ctx models.Context) { ctx.RespondString("Hjh") })

	return ws
}

func TestWebServerCoreRoutes(t *testing.T) {
	// Spin up the core test server instance
	ws := setupBaseServer(4041)
	go func() {
		_ = ws.ListenHttp()
	}()
	time.Sleep(50 * time.Millisecond) // Give the port a moment to bind

	t.Run("GET /test1", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:4041/test1")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "hello from test1" {
			t.Errorf("Expected 'hello from test1', got '%s'", string(body))
		}
	})

	t.Run("POST /test1 with body", func(t *testing.T) {
		jsonBody := []byte(`{"ping": "pong"}`)
		resp, err := http.Post("http://127.0.0.1:4041/test1", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "hello from test1 post" {
			t.Errorf("Expected payload signature mismatch, got '%s'", string(body))
		}
	})

	t.Run("GET /test2 JSON Marshalling", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:4041/test2")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("Expected header Content-Type application/json, got '%s'", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("Group Routes v2/hallo", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:4041/v2/hallo")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "H" {
			t.Errorf("Expected 'H', got '%s'", string(body))
		}
	})

	t.Run("Nested Group Wildcards v2/v1/abc/23/hallo", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:4041/v2/v1/abc/23/hallo")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "Hjh" {
			t.Errorf("Expected 'Hjh' from dynamic group mapping, got '%s'", string(body))
		}
	})
}

func TestWebServerMiddlewareLifecycle(t *testing.T) {
	ws := webServer.NewWebServer("127.0.0.1", 4042)
	ws.SetLogLevel(log.DebugLevel)

	g := ws.Group("v2")
	tz := false // State variable evaluated inside the request scope closure

	m := models.Middleware(func(hf models.HandlerFunc) models.HandlerFunc {
		return func(ctx models.Context) {
			if !tz {
				ctx.RespondString("blocked")
				return
			}
			hf(ctx)
		}
	})

	g.Get("middleware", m(func(ctx models.Context) {
		ctx.RespondString("good middleware")
	}))

	go func() {
		_ = ws.ListenHttp()
	}()
	time.Sleep(50 * time.Millisecond)

	t.Run("Request Interception when tz=false", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:4042/v2/middleware")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "blocked" {
			t.Errorf("Middleware failed to drop context frame, returned: '%s'", string(body))
		}
	})

	t.Run("Pass-through execution when state toggles true", func(t *testing.T) {
		tz = true // Dynamically change the flag variable inside the closure space

		resp, err := http.Get("http://127.0.0.1:4042/v2/middleware")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "good middleware" {
			t.Errorf("Middleware failed to pass frame down pipeline, returned: '%s'", string(body))
		}
	})
}
