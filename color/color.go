package color

import "strings"

const (
	GET       string = "#6415f7"
	POST      string = "#56f8ba"
	OPTION    string = "#56f8ba"
	WEBSOCKET string = "#ff8800"
	ERROR     string = "#ff0000"
	ROUTER    string = "#042180"
)

func GetColor(name string) (s string) {
	switch {
	case strings.Contains(name, "GET"):
		s = GET
	case strings.Contains(name, "POST"):
		s = POST
	case strings.Contains(name, "OPTION"):
		s = OPTION
	case strings.Contains(name, "WEBSOCKET"), strings.Contains(name, "WS"):
		s = WEBSOCKET
	case strings.Contains(name, "ERROR"), strings.Contains(name, "ERR"), name == "CORS":
		s = ERROR
	case strings.Contains(name, "ROUTER"):
		s = ROUTER
	default:
		s = "#f0f17b"
	}

	return
}
