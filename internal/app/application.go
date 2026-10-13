package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yogavredizon/log-analyzer/internal/bootstrap"
	"github.com/yogavredizon/log-analyzer/internal/config"
	"github.com/yogavredizon/log-analyzer/internal/sdk/rabbitmq"
)

type Application struct {
	server        *http.Server
	rabbitConn    *rabbitmq.Connection
	traceShutdown func(ctx context.Context)
}

func NewApplication(cfg *config.Config) (*Application, error) {
	log := bootstrap.NewLogger()

	tp, shutdownTracing, err := bootstrap.NewTracing(cfg)
	if err != nil {
		return nil, err
	}

	rabbitConn, publisher, err := bootstrap.NewRabbitMQ(cfg)
	if err != nil {
		shutdownTracing(context.Background())
		return nil, err
	}

	handler := bootstrap.NewHTTPHandler(cfg, log, tp, publisher)

	return &Application{
		server: &http.Server{
			Addr:         cfg.AppPort,
			Handler:      handler,
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
