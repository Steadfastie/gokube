package job

import (
	"context"
	"encoding/json"

	"github.com/steadfastie/gokube/data/brocker"
	"github.com/steadfastie/gokube/data/events"
	"github.com/steadfastie/gokube/data/repositories"
	"github.com/steadfastie/gokube/data/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type ConsumerProcessor interface {
	Process(ctx context.Context)
}

const collection = "counter"

type consumerProcessor struct {
	Collection *mongo.Collection
	Consumer   brocker.Consumer
	EventsRepo repositories.EventsRepository
	Logger     *zap.Logger
}

func NewConsumerProcessor(mongodb *services.MongoDB, consumer brocker.Consumer, repo repositories.EventsRepository, logger *zap.Logger) ConsumerProcessor {
	return &consumerProcessor{
		Collection: mongodb.MongoDB.Collection(collection),
		Consumer:   consumer,
		EventsRepo: repo,
		Logger:     logger,
	}
}

func (processor *consumerProcessor) Process(ctx context.Context) {
	processor.Logger.Info("Standing by for messages")
	messageChan := make(chan []byte)
	errChan := make(chan error)

	defer close(messageChan)
	defer close(errChan)

	go processor.Consumer.RecieveMessage(ctx, messageChan, errChan)

	select {
	case message := <-messageChan:
		var event events.CounterEvent
		if err := json.Unmarshal(message, &event); err != nil {
			processor.Logger.Error("Consumer could not recognize message", zap.Error(err))
		}
		processor.Logger.Info("Received message", zap.Any("event", event))
		go processor.EventsRepo.SaveEvent(ctx, &event)
	case err := <-errChan:
		processor.Logger.Error("Could not receive messages", zap.Error(err))
	}
}
