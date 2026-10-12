package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

type traceIdFn func(context.Context) string

type Logger struct {
	discard   bool
	handler   slog.Handler
	traceIdFn traceIdFn
}

func New(w io.Writer, level Level, service string, traceIdFn traceIdFn) *Logger {
	return new(w, level, service, traceIdFn, Events{})
}

func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelInfo, 3, msg, args...)
}

func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelError, 3, msg, args...)
}

func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelWarn, 3, msg, args...)
}

func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelDebug, 3, msg, args...)
}

func (log *Logger) write(ctx context.Context, level Level, caller int, msg string, args ...any) {
	logLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, logLevel) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), logLevel, msg, pcs[0])

	if log.traceIdFn != nil {
		args = append(args, log.traceIdFn(ctx))
	}

	log.handler.Handle(ctx, r)
}

func new(w io.Writer, miniLevel Level, service string, traceId traceIdFn, events Events) *Logger {
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: slog.Level(miniLevel), ReplaceAttr: f}))

	if events.Debug != nil || events.Error != nil || events.Info != nil || events.Warn != nil {
		handler = newLogHandler(handler, events)
	}

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(service)},
	}

	handler = handler.WithAttrs(attrs)

	return &Logger{
		handler:   handler,
		traceIdFn: traceId,
	}
}
