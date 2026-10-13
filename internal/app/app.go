package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/yogavredizon/project-template/internal/config"
	"github.com/yogavredizon/project-template/internal/http/middleware"
	"github.com/yogavredizon/project-template/internal/http/router"
	"github.com/yogavredizon/project-template/internal/module/activity"
	"github.com/yogavredizon/project-template/internal/sdk/rabbitmq"
	"github.com/yogavredizon/project-template/pkg/logger"
	"github.com/yogavredizon/project-template/pkg/otel"
	"github.com/yogavredizon/project-template/pkg/web"
)

type Application struct {
	server        *http.Server
	rabbitConn    *rabbitmq.Connection
	traceShutdown func(ctx context.Context)
}

func NewApplication(cfg *config.Config) (*Application, error) {
	otelCfg := otel.Config{
		Host:        cfg.OtelHost,
		ServiceName: cfg.AppName,
		Probability: 1.0,
	}

	tp, shutdownTracing, err := otel.InitTracing(&otelCfg)
	if err != nil {
		return nil, err
	}

	log := logger.New(os.Stdout, logger.LevelDebug, "main", func(context.Context) string {
		return uuid.New().String()
	})

	middlewares := []web.MidFun{
		middleware.Logger(log), middleware.Error(log),
	}

	app := web.NewApp(log.Info, tp.Tracer(cfg.AppName), middlewares...)
	app.EnableOrigin([]string{"*"})

	rabbitConn := rabbitmq.NewConnection(cfg)
	publisher, err := rabbitmq.NewPublisher(rabbitConn)
	if err != nil {
		rabbitConn.Shutdown()
		shutdownTracing(context.Background())
		return nil, err
	}

	activityModule := activity.NewModule(publisher)
	router.New(app, activityModule)

	return &Application{
		server: &http.Server{
			Addr:         ":8888",
			Handler:      app,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		rabbitConn:    rabbitConn,
		traceShutdown: shutdownTracing,
	}, nil
}

func (a *Application) Start() error {
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *Application) Shutdown(ctx context.Context) error {
	err := a.server.Shutdown(ctx)
	a.rabbitConn.Shutdown()
	a.traceShutdown(ctx)
	return err
}
