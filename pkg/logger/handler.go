package logger

import (
	"context"
	"log/slog"
)

type logHandler struct {
	events  Events
	handler slog.Handler
}

func newLogHandler(h slog.Handler, e Events) *logHandler {
	return &logHandler{events: e, handler: h}
}

func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{handler: h.handler.WithAttrs(attrs), events: h.events}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{handler: h.handler.WithGroup(name), events: h.events}
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		if h.events.Debug != nil {
			h.events.Debug(ctx, record(r))
		}
	case slog.LevelInfo:
		if h.events.Info != nil {
			h.events.Info(ctx, record(r))
		}
	case slog.LevelWarn:
		if h.events.Warn != nil {
			h.events.Warn(ctx, record(r))
		}
	case slog.LevelError:
		if h.events.Error != nil {
			h.events.Error(ctx, record(r))
		}
	}

	return h.handler.Handle(ctx, r)
}
