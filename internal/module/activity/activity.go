package activity

import (
	"net/http"

	"github.com/yogavredizon/project-template/internal/module/activity/handler"
	"github.com/yogavredizon/project-template/internal/module/activity/service"
	"github.com/yogavredizon/project-template/internal/sdk/rabbitmq"
	"github.com/yogavredizon/project-template/pkg/web"
)

type Module struct {
	Handler *handler.ActivityHandler
}

func NewModule(publisher *rabbitmq.Publisher) *Module {
	service := service.NewActivity(publisher)
	handler := handler.NewActivityHandler(service)

	return &Module{
		Handler: handler,
	}
}

func (m *Module) Route(app *web.App) {
	app.HandlerFunc(http.MethodPost, "activities", "/publish", m.Handler.Publish)
}
