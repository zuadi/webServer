package utils

import (
	"net"
	"strings"
)

func CleanPath(p string) string {
	if p == "/" {
		return "/"
	}
	return "/" + strings.Trim(p, "/")
}

func GetFreePort(address string, port int) (int, error) {
	if port != 0 {
		return port, nil
	}

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(address, "0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetCookie(cookie []string, key string) string {
	for _, element := range cookie {
		str := strings.Split(element, "=")
		if str[0] == key && len(str) > 1 {
			return str[1]
		}
	}
	return ""
}
