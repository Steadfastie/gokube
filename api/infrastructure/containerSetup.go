package infrastructure

import (
	"context"
	"log"

	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/api/handlers"
	"github.com/steadfastie/gokube/data/repositories"
	"github.com/steadfastie/gokube/data/services"
	"go.uber.org/zap"
)

func InitializeServices(ctx context.Context, logger *zap.Logger) {
	err := container.Singleton(func() (*Config, error) {
		return NewConfig(ctx, logger)
	})
	if err != nil {
		log.Fatalf("can't register configuration as singleton: %v", err)
	}

	err = container.Singleton(func(config *Config) (*zap.Logger, error) {
		return services.NewLogger(ctx, config)
	})
	if err != nil {
		log.Fatalf("can't register zap logger: %v", err)
	}

	err = container.Singleton(func(config *Config, logger *zap.Logger) (*services.MongoDB, error) {
		return services.NewMongoClient(ctx, config, logger)
	})
	if err != nil {
		log.Fatalf("can't register MongoDB client: %v", err)
	}

	err = container.Singleton(func(mongodb *services.MongoDB, logger *zap.Logger) repositories.CounterRepository {
		return repositories.NewCounterRepository(mongodb, logger)
	})
	if err != nil {
		log.Fatalf("can't register Basic repo: %v", err)
	}
}

func DisconnectServices(ctx context.Context) {
	container.Call(func(mongodb *services.MongoDB, logger *zap.Logger) {
		mongodb.DisconnectMongoClient(ctx, logger)
	})
	container.Call(func(logger *zap.Logger) {
		services.SyncLogger(logger)
	})
}

func GetCounterController() *handlers.CounterController {
	var controller handlers.CounterController
	container.Fill(&controller)
	return &controller
}
