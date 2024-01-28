package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/steadfastie/gokube/data"
	domainErrors "github.com/steadfastie/gokube/data/errors"
	"github.com/steadfastie/gokube/data/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const collection = "counter"

type CounterRepository interface {
	GetById(ctx context.Context, id string, resultChan chan<- *data.CounterDocument, errChan chan<- error)
	Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error)
	Patch(ctx context.Context, id string, patch *data.PatchModel, resultChan chan<- *data.PatchCounterResponse, errChan chan<- error)
}

type counterRepository struct {
	Collection *mongo.Collection
	Logger     *zap.Logger
}

func NewCounterRepository(mongodb *services.MongoDB, logger *zap.Logger) CounterRepository {
	return &counterRepository{
		Collection: mongodb.MongoDB.Collection(collection),
		Logger:     logger,
	}
}

func (repo *counterRepository) GetById(ctx context.Context, id string, resultChan chan<- *data.CounterDocument, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	var result struct {
		Document data.CounterDocument `bson:"document"`
	}
	filter := bson.M{"_id": objectID}
	opts := options.FindOne().SetProjection(bson.D{{Key: "document", Value: 1}})

	if err := repo.Collection.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewNotFoundError("Counter", id))
		}
		errChan <- err
	}
	resultChan <- &result.Document
}

func (repo *counterRepository) Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error) {
	now := time.Now().UTC()
	counterDocument := data.NewCounterDocument(now)
	document := data.NewDocument(counterDocument, counterDocument.Id)
	event := data.NewCounterCreatedEvent(ctx.Value("user").(string))
	outbox := data.NewOutboxEvent(event, now)

	document.Outbox.AddEvent(outbox)

	result, err := repo.Collection.InsertOne(ctx, document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", counterDocument.Id.Hex())))
		}
		repo.Logger.Error("Could not create document", zap.Any("Document", counterDocument), zap.Error(err))
		errChan <- err
	} else {
		resultChan <- result.InsertedID.(primitive.ObjectID)
	}
}

func (repo *counterRepository) Patch(ctx context.Context, id string, patch *data.PatchModel, resultChan chan<- *data.PatchCounterResponse, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	retryConfig := &data.RetryConfig{
		Context:          ctx,
		Logger:           repo.Logger,
		RecoverableError: mongo.ErrNoDocuments,
	}

	err = data.WithRetry(retryConfig, func() error {
		return repo.findOneAndUpdate(ctx, objectID, patch, resultChan)
	})

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Could not update document %v due to high service load", id)))
		}
		errChan <- err
	}
}

func (repo *counterRepository) findOneAndUpdate(ctx context.Context, id primitive.ObjectID, patch *data.PatchModel, resultChan chan<- *data.PatchCounterResponse) error {
	var counterBefore struct {
		Document data.CounterDocument `bson:"document"`
	}

	filter := bson.M{"_id": id}
	opts := options.FindOne().SetProjection(bson.D{{Key: "document", Value: 1}})

	if err := repo.Collection.FindOne(ctx, filter, opts).Decode(&counterBefore); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewNotFoundError("Counter", id.Hex()))
		}
		return err
	}

	var counterUpdate = counterBefore.Document.Copy()
	if patch.Increase {
		counterUpdate.IncreaseCounter(patch.UpdatedBy)
	} else {
		counterUpdate.DecreaseCounter(patch.UpdatedBy)
	}

	var counterAfter struct {
		Document data.CounterDocument `bson:"document"`
	}

	now := time.Now().UTC()
	event := data.NewCounterUpdatedEvent(counterUpdate.Counter, counterUpdate.UpdatedBy, ctx.Value("user").(string))
	outbox := data.NewOutboxEvent(event, now)

	updateFilter := bson.D{{Key: "_id", Value: counterUpdate.Id}, {Key: "document.version", Value: counterBefore.Document.Version}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "document.counter", Value: counterUpdate.Counter}}},
		{Key: "$set", Value: bson.D{{Key: "document.updatedAt", Value: now}}},
		{Key: "$set", Value: bson.D{{Key: "document.updatedBy", Value: counterUpdate.UpdatedBy}}},
		{Key: "$inc", Value: bson.D{{Key: "document.version", Value: 1}}},
		{Key: "$addToSet", Value: bson.D{{Key: "outbox.events", Value: outbox}}},
	}
	options := options.FindOneAndUpdate().SetProjection(bson.D{{Key: "document", Value: 1}}).SetReturnDocument(options.After)

	if err := repo.Collection.FindOneAndUpdate(ctx, updateFilter, update, options).Decode(&counterAfter); err != nil {
		return err
	}
	resultChan <- data.CreatePatchCounterResponse(&counterBefore.Document, &counterAfter.Document)
	return nil
}
