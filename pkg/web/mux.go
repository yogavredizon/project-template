package web

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Encode returning byte of data, content-type and error
type Encoder interface {
	Encode() ([]byte, string, error)
}

type HandlerFunc func(context.Context, *http.Request) Encoder

type Logger func(context.Context, string, ...any)

type App struct {
	log              Logger
	origins          []string
	otelMux          http.Handler
	mux              *http.ServeMux
	globalMiddleware []MidFun
	tracer           trace.Tracer
}

func NewApp(log Logger, tracer trace.Tracer, middlewares ...MidFun) *App {
	mux := http.NewServeMux()
	return &App{
		log:              log,
		mux:              mux,
		otelMux:          otelhttp.NewHandler(mux, "request"),
		globalMiddleware: middlewares,
		tracer:           tracer,
		origins:          []string{},
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.origins != nil {
		reqOrigin := r.Header.Get("Origin")
		for _, origin := range a.origins {
			if origin == "*" || origin == reqOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.Header().Set("Strict-Transport-Policy", "max-age=63072000,includeSubDomains,preload")
	r.URL.Path = path.Clean(r.URL.Path)

	a.otelMux.ServeHTTP(w, r)
}

func (a *App) HandlerFunc(method string, group, path string, handlerFunc HandlerFunc, middles ...MidFun) {
	handlerFunc = wrapMiddleware(middles, handlerFunc)
	handlerFunc = wrapMiddleware(a.globalMiddleware, handlerFunc)

	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "tracer", a.tracer)

		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

		resp := handlerFunc(ctx, r)

		if err := NewResponse(ctx, w, r, resp); err != nil {
			a.log(ctx, "web.service", "ERROR", err)
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, handler)
}

func (a *App) EnableOrigin(origins []string) {
	if origins != nil {
		a.origins = origins
	}
}
