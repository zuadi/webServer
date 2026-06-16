package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/log"
	"github.com/zuadi/webServer"
	"github.com/zuadi/webServer/models"
)

func main() {
	s := webServer.NewWebServer("0.0.0.0", 3333)
	s.SetLogLevel(log.DebugLevel)

	s.Get("info", func(ctx models.Context) {
		ctx.RespondString("Server running")
	})

	ws := s.NewWebSocket("ws")

	ws.Listen(func(data any) {
		if d, ok := data.([]byte); ok {
			var jsonData map[string]any
			err := json.Unmarshal(d, &jsonData)
			if err != nil {
				ws.Answer(1, []byte(`{ "error": ""`+err.Error()+`"}`))
			}
			var c string
			var ar []string

			if v, ok := jsonData["args"]; ok {
				switch val := v.(type) {
				case []any:
					for _, a := range val {
						ar = append(ar, fmt.Sprint(a))
					}
				case []string:
					for _, a := range val {
						ar = append(ar, a)
					}
				case string:
					ar = append(ar, val)
				}
			}
			if v, ok := jsonData["cmd"]; ok {
				c = fmt.Sprint(v)
			}

			cmd := exec.Command(c, ar...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				ws.Answer(1, []byte(`{ "error": ""`+err.Error()+`"}`))
				return
			}

			ws.Answer(1, output)
			return
		}

		ws.Answer(1, []byte("websocket data is not []byte"))
	})

	b, err := s.NewBroker("mqtt/pub/*", "0.0.0.0", 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	err = b.Subscribe("adrian/#", func(msg models.MQTTMessage) {
		fmt.Println(222, msg)
	})
	if err != nil {
		log.Fatal(err)
	}

	s.ListenHttp()
}
