package router

import "github.com/yogavredizon/project-template/pkg/web"

type Router interface {
	Route(app *web.App)
}

func New(app *web.App, routes ...Router) {
	for _, route := range routes {
		route.Route(app)
	}
}
