package logger

import (
	"context"
	"log/slog"
	"time"
)

type Level slog.Level

const (
	LevelDebug = Level(slog.LevelDebug)
	LevelWarn  = Level(slog.LevelWarn)
	LevelInfo  = Level(slog.LevelInfo)
	LevelError = Level(slog.LevelError)
)

type Record struct {
	Level      Level          `json:"level"`
	Time       time.Time      `json:"time"`
	Message    string         `json:"message"`
	Attributes map[string]any `json:"attributes"`
}

func record(r slog.Record) Record {
	attrs := make(map[string]any, r.NumAttrs())

	f := func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	}

	r.Attrs(f)

	return Record{
		Time:       r.Time,
		Message:    r.Message,
		Level:      Level(r.Level),
		Attributes: attrs,
	}
}

type EventFn func(ctx context.Context, r Record)

type Events struct {
	Debug EventFn
	Info  EventFn
	Warn  EventFn
	Error EventFn
}
