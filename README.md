# Project Template

A small Go HTTP service template with:
- CLI entrypoints in `cmd/`
- application wiring in `internal/app/`
- configuration in `internal/config/`
- HTTP middleware and routing in `internal/http/`
- feature modules in `internal/module/`
- reusable packages in `pkg/`
- RabbitMQ publishing integration
- OpenTelemetry tracing
- structured logging and HTTP middleware

## Project layout

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── config/
│   │   └── config.go
│   ├── http/
│   │   ├── middleware/
│   │   │   ├── errors.go
│   │   │   ├── helper.go
│   │   │   └── logger.go
│   │   └── router/
│   │       └── router.go
│   ├── module/
│   │   └── activity/
│   │       ├── activity.go
│   │       ├── dto/
│   │       ├── handler/
│   │       └── service/
│   └── sdk/
│       └── rabbitmq/
│           ├── connection.go
│           ├── consumer.go
│           ├── message.go
│           └── publisher.go
├── pkg/
│   ├── errs/
│   │   ├── codes.go
│   │   └── http_error.go
│   ├── logger/
│   │   ├── handler.go
│   │   ├── logger.go
│   │   └── model.go
│   ├── otel/
│   │   ├── otel.go
│   │   └── sampler.go
│   └── web/
│       ├── middleware.go
│       ├── mux.go
│       ├── request.go
│       └── response.go
├── .gitignore
├── .zarf/
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── .env.example
```

## Run the app

1. Copy `.env.example` to `.env` and fill in your values.
2. Start dependencies if needed (RabbitMQ and OTEL collector).
3. Run:

```bash
go run ./cmd/api
```

## Configuration

The application reads environment variables via Viper. Example values:

```env
APP_NAME=project-template
OTEL_HOST=localhost:4317
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_MAX_ATTEMPTS=5
RABBITMQ_RECONNECT_DELAY=3
RABBITMQ_CHANNEL_POOL_SIZE=20
RABBITMQ_PUBLISH_TIMEOUT=5
```

## Notes

- `internal/` is reserved for app-specific code that should not be imported outside the module.
- `pkg/` is for reusable packages, but should stay intentionally generic and not contain service-local logic.
- The app is structured for a small backend service and is easy to extend with more modules under `internal/module`.
