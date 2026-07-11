package models

import (
	"github.com/zuadi/webServer/constants"
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

func (g *Group) Get(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_GET, path)
	g.Route.Insert(constants.METHOD_GET, path, handler)
}

func (g *Group) Post(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_POST, path)
	g.Route.Insert(constants.METHOD_POST, path, handler)
}

func (g *Group) Put(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_PUT, path)
	g.Route.Insert(constants.METHOD_PUT, path, handler)
}

func (g *Group) Update(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_UPDATE, path)
	g.Route.Insert(constants.METHOD_UPDATE, path, handler)
}

func (g *Group) Delete(path string, handler HandlerFunc) {
	path = utils.CleanPath(g.Path) + utils.CleanPath(path)
	logger.DebugWithStyle(constants.METHOD_DELETE, path)
	g.Route.Insert(constants.METHOD_DELETE, path, handler)
}
