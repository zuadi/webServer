package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
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
	g := ws.NewGroup("v2")
	g.Get("hallo", func(ctx models.Context) { ctx.RespondString("H") })
	g.Get("velo", func(ctx models.Context) { ctx.RespondString("hallo velo") })

	g2 := g.NewGroup("v1")
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

	g := ws.NewGroup("v2")
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

func TestWebServerWebSocket(t *testing.T) {
	ws := webServer.NewWebServer("127.0.0.1", 4043)
	ws.SetLogLevel(log.DebugLevel)

	ws.Get("/ws", func(ctx models.Context) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		// Upgrade the HTTP server connection to the WebSocket protocol
		// NOTE: Replace .Writer() and .Request() with your framework's actual getter methods
		conn, err := upgrader.Upgrade(ctx.GetResponseWriter(), ctx.GetRequest(), nil)
		if err != nil {
			log.Errorf("Upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// Core echo loop
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break // connection closed by client or error
			}
			// Echo the exact message back
			if err := conn.WriteMessage(mt, message); err != nil {
				break
			}
		}
	})

	go func() {
		_ = ws.ListenHttp()
	}()
	time.Sleep(50 * time.Millisecond)

	t.Run("WebSocket Echo Connection", func(t *testing.T) {
		u := "ws://127.0.0.1:4043/ws"
		dialer := websocket.DefaultDialer
		conn, _, err := dialer.Dial(u, nil)
		if err != nil {
			t.Fatalf("Handshake failed: %v", err)
		}
		defer conn.Close()

		input := []byte("ping")
		if err := conn.WriteMessage(websocket.TextMessage, input); err != nil {
			t.Fatalf("Failed to write message: %v", err)
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		if string(message) != "ping" {
			t.Errorf("Expected 'ping', got '%s'", string(message))
		}
	})
}

func TestWebServer(t *testing.T) {
	ws := webServer.NewWebServer("localhost", 4040)
	ws.SetLogLevel(log.DebugLevel)

	// Register WebSocket & File endpoints
	ws.NewWebSocket("/ws")
	ws.ServeFile("/testserver", "../mocks/index.html")
	ws.ServeFileSystem("/getesten/*", "../models")

	// Register Route Handlers
	ws.Get("/test1", func(ctx models.Context) { ctx.RespondString("hello from test1") })
	ws.Post("/test1", func(ctx models.Context) { ctx.RespondString("hello from test1") })

	ws.Get("/test2", func(ctx models.Context) {
		data := struct {
			Info    string `json:"Info"`
			Message string `json:"Message"`
		}{
			Info:    "OK",
			Message: "This is a message",
		}
		ctx.RespondJson(http.StatusOK, data)
	})

	// Group Route Handlers
	g := ws.NewGroup("v2")
	g.Get("hallo", func(ctx models.Context) { ctx.RespondString("H") })
	g.Get("velo", func(ctx models.Context) { ctx.RespondString("hallo velo") })

	g2 := g.NewGroup("v1")
	g2.Get("/:id/23/hallo", func(ctx models.Context) { ctx.RespondString("Hjh") })

	// Start server asynchronously
	go func() {
		if err := ws.ListenHttp(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	go func() {
		_ = ws.ListenHttp()
	}()
	time.Sleep(50 * time.Millisecond)

	webocket := g.NewWebSocket("ws")

	webocket.NewConnection = func(ws *models.WSClient) {
		fmt.Println(1, "New conection")
		ws.Answer(1, []byte("Hello WebSocket"))
	}

	// Allow server time to bind to the port
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://localhost:4040"

	// ==========================================
	// SUBTESTS
	// ==========================================

	t.Run("GET /test1", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/test1")
		if err != nil {
			t.Fatalf("Failed to execute GET request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "hello from test1" {
			t.Errorf("Expected 'hello from test1', got '%s'", string(body))
		}
	})

	t.Run("POST /test1", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/test1", "text/plain", nil)
		if err != nil {
			t.Fatalf("Failed to execute POST request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /v2/velo", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v2/velo")
		if err != nil {
			t.Fatalf("Failed to execute GET request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "hallo velo" {
			t.Errorf("Expected 'hallo velo', got '%s'", string(body))
		}
	})

	t.Run("GET /v2/v1/123/23/hallo", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v2/v1/123/23/hallo")
		if err != nil {
			t.Fatalf("Failed to execute GET request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Hjh" {
			t.Errorf("Expected 'Hjh', got '%s'", string(body))
		}
	})

	t.Run("GET /v2/v1/123/23/hallo", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/v2/v1/123/23/hallo")
		if err != nil {
			t.Fatalf("Failed to execute GET request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Hjh" {
			t.Errorf("Expected 'Hjh', got '%s'", string(body))
		}
	})

	t.Run("WS /v2/ws", func(t *testing.T) {
		// Convert http://localhost:4040 to ws://localhost:4040/v2/ws
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/v2/ws"

		// 1. Establish the WebSocket connection
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket at %s: %v", wsURL, err)
		}
		defer conn.Close()

		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("Expected status 101 Switching Protocols, got %d", resp.StatusCode)
		}

		// 2. Write a message to the WebSocket
		testMessage := []byte("Hello WebSocket")
		err = conn.WriteMessage(websocket.TextMessage, testMessage)
		if err != nil {
			t.Fatalf("Failed to send WebSocket message: %v", err)
		}

		// 3. Read the response/echo back
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read WebSocket message: %v", err)
		}

		if messageType != websocket.TextMessage {
			t.Errorf("Expected TextMessage (%d), got %d", websocket.TextMessage, messageType)
		}

		if string(p) != "Hello WebSocket" {
			t.Errorf("Expected 'Hello WebSocket', got '%s'", string(p))
		}
	})
}
