package test

import (
	"context"
	"net/http"
	"testing"
	"time"
	"webServer"
	"webServer/models"
)

func TestWebServer(t *testing.T) {
	t.Log("start webserver test")

	var timeout time.Duration
	timeout = 120 * time.Second
	t.Logf("set test timeout to: %v seconds", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ws := webServer.NewWebServer("0.0.0.0", 4040)

	ws.Get("/test1", func(ctx models.Context) { ctx.RespondString("hello from test1") })
	ws.ServeFile("/testserver", "../index.html")
	ws.ServeFileSystem("/getesten/*", "../models")

	ws.Get("/test2", func(ctx models.Context) {
		var data struct {
			Info    string
			Message string
		}

		data.Info = "OK"
		data.Message = "This is a message"
		ctx.RespondJson(http.StatusOK, data)
	})

	g := ws.Group("v2")

	g.Get("hallo", func(ctx models.Context) { ctx.RespondString("H") })
	g.Get("velo", func(ctx models.Context) {
		ctx.RespondString("hallo velo")
	})
	//g.Get("/:id/23/hallo", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hjh")) })
	g2 := g.Group("v1")
	g2.Get("/:id/23/hallo", func(ctx models.Context) { ctx.RespondString("Hjh") })

	go func() {
		for {
			if ctx.Err() != nil {
				t.Logf("test finished after %v seconds timeout", timeout)
				break
			}
		}
	}()

	err := ws.ListenHttp()
	if err != nil {
		t.Fatal(err)
	}

}
