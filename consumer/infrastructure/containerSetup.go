package infrastructure

import (
	"context"
	"log"

	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/consumer/job"
	"github.com/steadfastie/gokube/data/brocker"
	"github.com/steadfastie/gokube/data/repositories"
	"github.com/steadfastie/gokube/data/services"
	"go.uber.org/zap"
)

func InitializeServices(ctx context.Context, logger *zap.Logger) {
	err := container.Singleton(func() (*Config, error) {
		return NewConfig()
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

	err = container.Singleton(func(config *Config, logger *zap.Logger) brocker.Consumer {
		return brocker.NewConsumer(ctx, logger, config.KafkaServers...)
	})
	if err != nil {
		log.Fatalf("can't register Basic repo: %v", err)
	}

	err = container.Singleton(func(mongodb *services.MongoDB, logger *zap.Logger) repositories.EventsRepository {
		return repositories.NewEventsRepository(mongodb, logger)
	})
	if err != nil {
		log.Fatalf("can't register Basic repo: %v", err)
	}

	err = container.Singleton(func(mongodb *services.MongoDB, consumer brocker.Consumer, repo repositories.EventsRepository, logger *zap.Logger) job.ConsumerProcessor {
		return job.NewConsumerProcessor(mongodb, consumer, repo, logger)
	})
	if err != nil {
		log.Fatalf("can't register Basic repo: %v", err)
	}
}

func DisconnectServices(ctx context.Context) {
	container.Call(func(consumer brocker.Consumer) {
		consumer.Disconnect()
	})
	container.Call(func(mongodb *services.MongoDB, logger *zap.Logger) {
		mongodb.DisconnectMongoClient(ctx, logger)
	})
	container.Call(func(logger *zap.Logger) {
		services.SyncLogger(logger)
	})
}

func GetCron() string {
	var config *Config
	container.Resolve(&config)
	return config.Cron
}

func GetOutboxProcessor() job.ConsumerProcessor {
	var processor job.ConsumerProcessor
	container.Resolve(&processor)
	return processor
}
