package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yogavredizon/project-template/internal/app"
	"github.com/yogavredizon/project-template/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	application, err := app.NewApplication(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startErr := make(chan error, 1)
	go func() {
		startErr <- application.Start()
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-startErr:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return errors.Join(runErr, application.Shutdown(shutdownCtx))
}
