package color

import (
	"strings"

	"github.com/zuadi/webServer/constants"
)

func GetColor(name string) (s string) {
	switch {
	case strings.Contains(name, "GET"):
		s = constants.COLOR_GET
	case strings.Contains(name, "POST"):
		s = constants.COLOR_POST
	case strings.Contains(name, "PUT"):
		s = constants.COLOR_PUT
	case strings.Contains(name, "UPDATE"):
		s = constants.COLOR_UPDATE
	case strings.Contains(name, "DELETE"):
		s = constants.COLOR_DELETE
	case strings.Contains(name, "OPTION"):
		s = constants.COLOR_OPTION
	case strings.Contains(name, "WEBSOCKET"), strings.Contains(name, "WS"):
		s = constants.COLOR_WEBSOCKET
	case strings.Contains(name, "ERROR"), strings.Contains(name, "ERR"), name == "CORS":
		s = constants.COLOR_ERROR
	case strings.Contains(name, "ROUTER"):
		s = constants.COLOR_ROUTER
	default:
		s = "#f0f17b"
	}

	return
}
