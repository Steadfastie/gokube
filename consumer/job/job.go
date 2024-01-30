package job

import (
	"context"
	"encoding/json"

	"github.com/steadfastie/gokube/data/brocker"
	"github.com/steadfastie/gokube/data/events"
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
	Logger     *zap.Logger
}

func NewConsumerProcessor(mongodb *services.MongoDB, consumer brocker.Consumer, logger *zap.Logger) ConsumerProcessor {
	return &consumerProcessor{
		Collection: mongodb.MongoDB.Collection(collection),
		Consumer:   consumer,
		Logger:     logger,
	}
}

func (processor *consumerProcessor) Process(ctx context.Context) {
	processor.Logger.Info("Starting receiving messages")
	messageChan := make(chan []byte)
	errChan := make(chan error)

	defer close(messageChan)
	defer close(errChan)

	go processor.Consumer.RecieveMessage(ctx, messageChan, errChan)

	select {
	case message := <-messageChan:
		var event events.CounterEvent
		if err := json.Unmarshal(message, &event); err != nil {
			processor.Logger.Error("Consumer not recognize message", zap.Error(err))
		}
		processor.Logger.Info(
			"Event happened",
			zap.Any("event", event),
		)
	case <-errChan:
	}

	processor.Logger.Info("Completing receiving messages")
}
