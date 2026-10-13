package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/yogavredizon/project-template/pkg/errs"
	"github.com/yogavredizon/project-template/pkg/logger"
	"github.com/yogavredizon/project-template/pkg/web"
)

func Error(log *logger.Logger) web.MidFun {
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx context.Context, r *http.Request) web.Encoder {
			resp := next(ctx, r)

			log.Info(ctx, "Error middleware", resp)

			if err := checkIsError(resp); err != nil {
				errApp, ok := errors.AsType[*errs.HttpError](err)
				if ok {
					return errApp
				}
			}

			return resp
		}
	}
}
