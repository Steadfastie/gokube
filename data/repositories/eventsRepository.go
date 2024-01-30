package repositories

import (
	"context"
	"time"

	"github.com/steadfastie/gokube/data/events"
	"github.com/steadfastie/gokube/data/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const eventsCollection = "events"

type EventsRepository interface {
	SaveEvent(ctx context.Context, event *events.CounterEvent)
}

type eventsRepository struct {
	Collection *mongo.Collection
	Logger     *zap.Logger
}

func NewEventsRepository(mongodb *services.MongoDB, logger *zap.Logger) EventsRepository {
	return &eventsRepository{
		Collection: mongodb.MongoDB.Collection(eventsCollection),
		Logger:     logger,
	}
}

func (repo *eventsRepository) SaveEvent(ctx context.Context, event *events.CounterEvent) {
	event.AddTrail(events.Consumer, time.Now().UTC())

	_, err := repo.Collection.InsertOne(ctx, event)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			repo.Logger.Info("Event has been already consumed", zap.String("id", event.EventId.Hex()))
		}
		repo.Logger.Error("Error saving event", zap.String("id", event.EventId.Hex()), zap.Error(err))
	}
}
