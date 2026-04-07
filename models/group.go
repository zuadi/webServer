package models

import (
	"github.com/zuadi/webServer/utils"
)

type Group struct {
	Path  string
	Route *Route
}

func (g *Group) Group(path string) *Group {
	return &Group{
		Path:  g.Path + "/" + utils.CleanPath(path),
		Route: g.Route,
	}
}

func (g *Group) Get(path string, handler Handler) {
	g.Route.Insert("GET", g.Path+"/"+utils.CleanPath(path), handler)
}

func (g *Group) Post(path string, handler Handler) {
	g.Route.Insert("Post", g.Path+"/"+utils.CleanPath(path), handler)
}
