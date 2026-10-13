package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/yogavredizon/log-analyzer/pkg/logger"
	"github.com/yogavredizon/log-analyzer/pkg/otel"
	"github.com/yogavredizon/log-analyzer/pkg/web"
	"go.opentelemetry.io/otel/attribute"
)

func Logger(log *logger.Logger) web.MidFun {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx context.Context, r *http.Request) web.Encoder {
			now := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s:%s", path, r.URL.RawQuery)
			}

			log.Info(ctx, "request started", "method", r.Method, "path", path, "remoteAddr", r.RemoteAddr)

			ctx, span := otel.AddSpan(ctx, "middleware.logger", attribute.String("path", path))
			defer span.End()

			resp := next(ctx, r)

			var statusCode int
			if err := checkIsError(resp); err != nil {
				statusCode = http.StatusInternalServerError
			}

			log.Info(ctx, "request completed", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr, "statuscode", statusCode, "since", time.Since(now).String())

			return resp
		}
	}
}
