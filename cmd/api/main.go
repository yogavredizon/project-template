package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/yogavredizon/log-analyzer/internal/middleware"
	"github.com/yogavredizon/log-analyzer/pkg/errs"
	"github.com/yogavredizon/log-analyzer/pkg/logger"
	"github.com/yogavredizon/log-analyzer/pkg/otel"
	"github.com/yogavredizon/log-analyzer/pkg/web"
)

func main() {
	cfg := otel.Config{
		Host:        "172.17.151.164:4317",
		ServiceName: "infolelang",
		Probability: 1.0,
	}

	tp, shutdown, err := otel.InitTracing(&cfg)
	if err != nil {
		panic(err)
	}

	defer shutdown(context.Background())

	var log *logger.Logger = logger.New(os.Stdout, logger.LevelDebug, "main", func(ctx context.Context) string {
		return uuid.New().String()
	})

	middlewares := []web.MidFun{
		middleware.Logger(log), middleware.Error(log),
	}

	app := web.NewApp(log.Info, tp.Tracer(cfg.ServiceName), middlewares...)

	app.EnableOrigin([]string{"*"})

	app.HandlerFunc(http.MethodGet, "", "/healthz", health, nil)

	api := http.Server{
		Addr:         ":8888",
		Handler:      app,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := api.ListenAndServe(); err != nil {
		panic(err)
	}
}

func health(ctx context.Context, r *http.Request) web.Encoder {
	return errs.New(errs.ValidationError, "Ini error request")
}
