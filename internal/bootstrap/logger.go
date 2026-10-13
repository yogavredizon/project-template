package bootstrap

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/yogavredizon/log-analyzer/pkg/logger"
)

func NewLogger() *logger.Logger {
	return logger.New(os.Stdout, logger.LevelDebug, "main", func(context.Context) string {
		return uuid.New().String()
	})
}
