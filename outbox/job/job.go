package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/steadfastie/gokube/data"
	"github.com/steadfastie/gokube/data/brocker"
	"github.com/steadfastie/gokube/data/events"
	"github.com/steadfastie/gokube/data/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type OutboxProcessor interface {
	ProcessOutbox(ctx context.Context)
}

const collection = "counter"

type outboxProcessor struct {
	Collection *mongo.Collection
	Producer   brocker.Producer
	Logger     *zap.Logger
}

func NewOutboxProcessor(mongodb *services.MongoDB, producer brocker.Producer, logger *zap.Logger) OutboxProcessor {
	return &outboxProcessor{
		Collection: mongodb.MongoDB.Collection(collection),
		Producer:   producer,
		Logger:     logger,
	}
}

func (processor *outboxProcessor) ProcessOutbox(ctx context.Context) {
	processor.Logger.Info("Starting processing")
	idsChan := make(chan []primitive.ObjectID)
	errChan := make(chan error)

	defer close(idsChan)
	defer close(errChan)

	go processor.findDocumentsToProcess(ctx, idsChan, errChan)

	var docIdsToHandle []primitive.ObjectID
	select {
	case ids := <-idsChan:
		if len(ids) == 0 {
			processor.Logger.Info("Outbox job found no events to handle")
			return
		}
		docIdsToHandle = ids
	case err := <-errChan:
		processor.Logger.Error("Outbox job caught error trying to parse collection", zap.Error(err))
		return
	}

	for _, id := range docIdsToHandle {
		go processor.handleEvents(ctx, id, errChan)

		err := <-errChan
		if err == nil {
			continue
		}
		processor.Logger.Error("Outbox job caught error trying handle event", zap.Error(err))
	}
	processor.Logger.Info("Completing processing")
}

func (processor *outboxProcessor) findDocumentsToProcess(ctx context.Context, resultChan chan<- []primitive.ObjectID, errChan chan<- error) {
	filter := bson.D{
		{Key: "outbox.events", Value: bson.D{{Key: "$ne", Value: bson.A{}}}},
	}
	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).SetLimit(20)

	cursor, err := processor.Collection.Find(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			resultChan <- nil
			return
		}
		errChan <- fmt.Errorf("error happened while finding documents with events: %w", err)
	}

	type documentWithId struct {
		Id primitive.ObjectID `bson:"_id"`
	}

	var results []documentWithId
	if err = cursor.All(ctx, &results); err != nil {
		errChan <- fmt.Errorf("error happened while pulling documents with events: %w", err)
	}

	stringResults := make([]primitive.ObjectID, len(results))
	for i, result := range results {
		stringResults[i] = result.Id
	}

	resultChan <- stringResults
}

func (processor *outboxProcessor) handleEvents(ctx context.Context, docId primitive.ObjectID, errChan chan<- error) {
	lockId := primitive.NewObjectID()
	lockChan := make(chan bool)
	defer close(lockChan)

	lockOptions := &LockOutboxOptions{
		ctx:        ctx,
		docId:      docId,
		lockId:     lockId,
		expiration: 30 * time.Second,
		resultChan: lockChan,
		errChan:    errChan,
	}
	go processor.lockOutbox(lockOptions)

	locked := <-lockChan
	if !locked {
		return
	}

	eventsChan := make(chan []data.OutboxEvent)
	defer close(eventsChan)

	go processor.getEvents(ctx, docId, eventsChan, errChan)

	events := <-eventsChan
	if len(events) == 0 {
		return
	}

	sort.Sort(data.ByTimestamp(events))

	for _, event := range events {
		eventChan := make(chan bool)
		defer close(eventChan)

		go processor.handleEvent(ctx, &event, eventChan, errChan)

		handled := <-eventChan
		if !handled {
			continue
		}

		// go processor.removeEvent(ctx, docId, event.Id, eventChan, errChan)
	}
	errChan <- nil
}

type LockOutboxOptions struct {
	ctx        context.Context
	docId      primitive.ObjectID
	lockId     primitive.ObjectID
	expiration time.Duration
	resultChan chan<- bool
	errChan    chan<- error
}

func (processor *outboxProcessor) lockOutbox(options *LockOutboxOptions) {
	now := time.Now().UTC()
	lockExpiration := now.Add(options.expiration)

	filter := bson.D{
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "_id", Value: options.docId}},
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "outbox.lockExpiration", Value: bson.D{{Key: "$eq", Value: nil}}}},
				bson.D{{Key: "outbox.lockExpiration", Value: bson.D{{Key: "$lte", Value: lockExpiration}}}},
				bson.D{{Key: "outbox.lockId", Value: bson.D{{Key: "$eq", Value: nil}}}},
				bson.D{{Key: "outbox.lockId", Value: bson.D{{Key: "$lte", Value: options.lockId}}}},
			},
			}},
		}},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "outbox.lockId", Value: options.lockId}}},
		{Key: "$set", Value: bson.D{{Key: "outbox.lockExpiration", Value: lockExpiration}}},
	}

	result, err := processor.Collection.UpdateOne(options.ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			options.resultChan <- false
			return
		}
		options.errChan <- fmt.Errorf("error happened locking outbox bucket for %v: %w", options.docId.Hex(), err)
	}

	options.resultChan <- result.ModifiedCount > 0
}

func (processor *outboxProcessor) getEvents(ctx context.Context, docId primitive.ObjectID, resultChan chan<- []data.OutboxEvent, errChan chan<- error) {
	filter := bson.M{"_id": docId}
	opts := options.FindOne().SetProjection(bson.D{{Key: "outbox", Value: 1}})

	var result struct {
		Outbox data.OutboxBucket `bson:"outbox"`
	}

	if err := processor.Collection.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			resultChan <- nil
			return
		}
		errChan <- fmt.Errorf("error happened while getting events: %w", err)
	}

	resultChan <- result.Outbox.Events
}

func (processor *outboxProcessor) handleEvent(ctx context.Context, event *data.OutboxEvent, resultChan chan bool, errChan chan<- error) {
	switch payload := event.Payload.(type) {
	case *data.CounterCreatedEvent:
		message := &events.CounterEvent{
			EventId: event.Id,
			When:    event.Timestamp,
			Who:     payload.UserAlias,
			What:    payload.Type,
		}
		value, err := json.Marshal(message)
		if err != nil {
			processor.Logger.Error("Error encoding event", zap.Error(err))
			return
		}
		processor.Logger.Info("Sending create counter event", zap.Any("Event", event))
		processor.Producer.SendMessage(ctx, []byte(string(payload.Type)), value)
	case *data.CounterUpdatedEvent:
		message := &events.CounterEvent{
			EventId: event.Id,
			When:    event.Timestamp,
			Who:     fmt.Sprintf("%v (%v)", payload.UpdatedBy, payload.UserAlias),
			What:    payload.Type,
		}
		value, err := json.Marshal(message)
		if err != nil {
			processor.Logger.Error("Error encoding event", zap.Error(err))
			return
		}
		processor.Logger.Info("Sending update counter event", zap.Any("Event", event))
		processor.Producer.SendMessage(ctx, []byte(string(payload.Type)), value)
	default:
		message := "Unknown event has been found. Won't ship that"
		processor.Logger.Info(message, zap.Any("Event", event))
	}

	resultChan <- true
}

func (processor *outboxProcessor) removeEvent(ctx context.Context, docId primitive.ObjectID, eventId primitive.ObjectID, resultChan chan bool, errChan chan<- error) {
	filter := bson.M{"_id": docId}
	update := bson.M{
		"$pull": bson.M{
			"outbox.events": bson.M{"_id": eventId},
		},
	}

	_, err := processor.Collection.UpdateOne(ctx, filter, update)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		errChan <- fmt.Errorf("error happened while removing events from %v: %w", docId.Hex(), err)
	}
}
