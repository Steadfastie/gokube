package infrastucture

import (
	"context"
	"log"

	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/data"
	"github.com/steadfastie/gokube/infrastucture/services"
	"go.uber.org/zap"
)

func InitializeServices(ctx context.Context, logger *zap.Logger) {
	err := container.Singleton(func() (*services.Config, error) {
		return services.NewConfig(ctx, logger)
	})
	if err != nil {
		log.Fatalf("can't register configuration as singleton: %v", err)
	}

	err = container.Singleton(func(config *services.Config) (*zap.Logger, error) {
		return services.NewLogger(ctx, config)
	})
	if err != nil {
		log.Fatalf("can't register zap logger: %v", err)
	}

	err = container.Singleton(func(config *services.Config, logger *zap.Logger) (*services.MongoDB, error) {
		return services.NewMongoClient(ctx, config, logger)
	})
	if err != nil {
		log.Fatalf("can't register MongoDB client: %v", err)
	}

	err = container.Singleton(func(mongodb *services.MongoDB, logger *zap.Logger) data.BasicRepository {
		return data.NewBasicRepository(mongodb, logger)
	})
	if err != nil {
		log.Fatalf("can't register Basic repo: %v", err)
	}
}

func DisconnectServices(ctx context.Context) {
	container.Call(func(mongodb *services.MongoDB, logger *zap.Logger) {
		mongodb.DisconnectMongoClient(ctx, logger)
	})
}
