package bootstrap

import (
	"net/http"

	"github.com/yogavredizon/log-analyzer/internal/config"
	"github.com/yogavredizon/log-analyzer/internal/http/middleware"
	router "github.com/yogavredizon/log-analyzer/internal/http/route"
	"github.com/yogavredizon/log-analyzer/internal/module/activity"
	"github.com/yogavredizon/log-analyzer/internal/sdk/rabbitmq"
	"github.com/yogavredizon/log-analyzer/pkg/logger"
	"github.com/yogavredizon/log-analyzer/pkg/web"
	"go.opentelemetry.io/otel/trace"
)

func NewHTTPHandler(cfg *config.Config, log *logger.Logger, tp trace.TracerProvider, publisher *rabbitmq.Publisher) http.Handler {
	middlewares := []web.MidFun{
		middleware.Logger(log),
		middleware.Error(log),
	}

	app := web.NewApp(log.Info, tp.Tracer(cfg.AppName), middlewares...)
	app.EnableOrigin([]string{"*"})

	activityModule := activity.NewModule(publisher)
	router.New(app, activityModule)

	return app
}
