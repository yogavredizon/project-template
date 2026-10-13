package bootstrap

import (
	"context"

	"github.com/yogavredizon/log-analyzer/internal/config"
	"github.com/yogavredizon/log-analyzer/pkg/otel"
	"go.opentelemetry.io/otel/trace"
)

func NewTracing(cfg *config.Config) (trace.TracerProvider, func(context.Context), error) {
	return otel.InitTracing(&otel.Config{
		Host:        cfg.OtelHost,
		ServiceName: cfg.AppName,
		Probability: 1.0,
	})
}
