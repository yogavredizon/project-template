package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yogavredizon/log-analyzer/internal/bootstrap"
	"github.com/yogavredizon/log-analyzer/internal/conf"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := conf.New()
	if err != nil {
		return err
	}

	app, err := bootstrap.NewApplication(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startErr := make(chan error, 1)
	go func() {
		startErr <- app.Start()
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-startErr:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return errors.Join(runErr, app.Shutdown(shutdownCtx))
}
