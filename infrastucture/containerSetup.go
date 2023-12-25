package infrastucture

import (
	"context"
	"github.com/golobby/container/v3"
	"go.uber.org/zap"
	"log"
)

func InitializeServices(ctx context.Context, logger *zap.Logger) {
	err := container.Singleton(func() (*Config, error) {
		return NewConfig(ctx, logger)
	})
	if err != nil {
		log.Fatalf("can't register configuration as singleton: %v", err)
	}

	err = container.Singleton(func(config *Config) (*zap.Logger, error) {
		return NewLogger(ctx, config)
	})
	if err != nil {
		log.Fatalf("can't register zap logger: %v", err)
	}
}
