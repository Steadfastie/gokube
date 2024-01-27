package job

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/steadfastie/gokube/data"
	"github.com/steadfastie/gokube/data/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const collection = "counter"

type outboxProcessor struct {
	Collection *mongo.Collection
	Logger     *zap.Logger
}

func ProcessOutbox(ctx context.Context, mongodb *services.MongoDB, logger *zap.Logger) {
	processor := &outboxProcessor{
		Collection: mongodb.MongoDB.Collection(collection),
		Logger:     logger,
	}
	idsChan := make(chan []primitive.ObjectID)
	errChan := make(chan error)

	go processor.findDocumentsToProcess(ctx, idsChan, errChan)

	var docIdsToHandle []primitive.ObjectID
	select {
	case ids := <-idsChan:
		if len(ids) == 0 {
			return
		}
		docIdsToHandle = ids
	case err := <-errChan:
		processor.Logger.Error("Outbox job caught error trying to parse collection", zap.Error(err))
		return
	}

	for _, id := range docIdsToHandle {
		go processor.handleEvents(ctx, id, errChan)

		select {
		case err := <-errChan:
			processor.Logger.Error("Outbox job caught error trying handle event", zap.Error(err))
			return
		}
	}
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
		errChan <- err
	}

	type documentWithId struct {
		Id primitive.ObjectID `bson:"document"`
	}

	var results []documentWithId
	if err = cursor.All(ctx, &results); err != nil {
		errChan <- err
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

	go processor.getEvents(ctx, docId, eventsChan, errChan)

	events := <-eventsChan
	if len(events) == 0 {
		return
	}

	sort.Sort(data.ByTimestamp(events))

	for _, event := range events {
		eventChan := make(chan bool)

		go processor.handleEvent(ctx, &event, eventChan, errChan)

		handled := <-eventChan
		if !handled {
			continue
		}

		go processor.removeEvent(ctx, docId, eventChan, errChan)

	}

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
		options.errChan <- err
	}

	options.resultChan <- result.ModifiedCount > 0
}

func (processor *outboxProcessor) getEvents(ctx context.Context, docId primitive.ObjectID, resultChan chan<- []data.OutboxEvent, errChan chan<- error) {
	filter := bson.M{"_id": docId}
	opts := options.FindOne().SetProjection(bson.D{{Key: "outbox.events", Value: 1}})

	var result struct {
		Events []data.OutboxEvent `bson:"events"`
	}

	if err := processor.Collection.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			resultChan <- nil
			return
		}
		errChan <- err
	}

	resultChan <- result.Events
}

func (processor *outboxProcessor) handleEvent(ctx context.Context, event *data.OutboxEvent, resultChan chan bool, errChan chan<- error) {

}

func (processor *outboxProcessor) removeEvent(ctx context.Context, docId primitive.ObjectID, resultChan chan bool, errChan chan<- error) {

}