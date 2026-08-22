package test

import (
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

	webocket := g.NewWebSocket("ws")

	webocket.NewConnection = func(ws *models.WSClient) {
		fmt.Println(1, "New conection")
		fmt.Println(2, ws)
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
