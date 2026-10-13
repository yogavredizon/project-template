package handler

import (
	"context"
	"net/http"

	"github.com/yogavredizon/log-analyzer/internal/module/activity/dto"
	"github.com/yogavredizon/log-analyzer/internal/module/activity/service"
	"github.com/yogavredizon/log-analyzer/pkg/errs"
	"github.com/yogavredizon/log-analyzer/pkg/web"
)

type ActivityHandler struct {
	service *service.Activity
}

func NewActivityHandler(service *service.Activity) *ActivityHandler {
	return &ActivityHandler{
		service: service,
	}
}

func (a *ActivityHandler) Publish(ctx context.Context, r *http.Request) web.Encoder {
	var req dto.Request
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.ValidationError, "invalid request")
	}

	if err := a.service.Publish(ctx, req); err != nil {
		return errs.New(errs.ValidationError, "invalid request")
	}

	return dto.Response{Status: "OK"}
}
