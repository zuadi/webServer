package color

const (
	GET       string = "#6415f7"
	POST      string = "#56f8ba"
	OPTION    string = "#56f8ba"
	WEBSOCKET string = "#ff8800"
	ERROR     string = "#ff0000"
)

func GetColor(name string) (s string) {
	switch name {
	case "GET":
		s = GET
	case "POST":
		s = POST
	case "OPTION":
		s = OPTION
	case "WEBSOCKET", "WS":
		s = WEBSOCKET
	case "ERROR", "ERR":
		s = ERROR
	default:
		s = "#f0f17b"
	}

	return
}
