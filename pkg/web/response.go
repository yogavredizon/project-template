package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/yogavredizon/project-template/pkg/errs"
	"github.com/yogavredizon/project-template/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
)

func NewResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, enc Encoder) error {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("client disconected, no request to send")
		}
	}

	code := http.StatusOK

	switch v := enc.(type) {
	case *errs.HttpError:
		code = int(v.Code)
	default:
		if enc == nil {
			code = http.StatusNoContent
		}
	}

	_, span := otel.AddSpan(ctx, "web.response", attribute.Int("status-code", code))
	defer span.End()

	data, contentType, err := enc.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
