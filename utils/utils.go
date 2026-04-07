package utils

import "strings"

func CleanPath(p string) string {
	if p == "/" {
		return "/"
	}
	return "/" + strings.Trim(p, "/")
}
