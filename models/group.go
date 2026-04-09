package models

import (
	"github.com/zuadi/webServer/logger"
	"github.com/zuadi/webServer/utils"
)

type Group struct {
	Path  string
	Route *Route
}

func (g *Group) Group(path string) *Group {
	return &Group{
		Path:  utils.CleanPath(g.Path) + utils.CleanPath(path),
		Route: g.Route,
	}
}

func (g *Group) Get(path string, handler Handler) {
	title := "GET"
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.SetStyle(title, "#56a7f8", path)
	g.Route.Insert(title, path, handler)
}

func (g *Group) Post(path string, handler Handler) {
	title := "POST"
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.SetStyle(title, "#56a7f8", path)
	g.Route.Insert(title, path, handler)
}
