package models

import (
	"strings"
)

type Group struct {
	Name  string
	Route *Route
}

func (g *Group) Group(name string) *Group {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")

	return &Group{
		Name:  g.Name + "/" + name,
		Route: g.Route,
	}
}

func (g *Group) Get(name string, handler Handler) {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")
	g.Route.Insert("GET", g.Name+"/"+name, handler)
}

func (g *Group) Post(name string, handler Handler) {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")
	g.Route.Insert("Post", g.Name+"/"+name, handler)
}
