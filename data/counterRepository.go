package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainErrors "github.com/steadfastie/gokube/infrastucture/errors"
	"github.com/steadfastie/gokube/infrastucture/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const collection = "counter"

type CounterRepository interface {
	GetById(ctx context.Context, id string, resultChan chan<- *CounterDocument, errChan chan<- error)
	Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error)
	Patch(ctx context.Context, document *CounterDocument, patchModel *PatchModel, resultChan chan<- *CounterDocument, errChan chan<- error)
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

func (repo *counterRepository) GetById(ctx context.Context, id string, resultChan chan<- *CounterDocument, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	var result struct {
		Document CounterDocument `bson:"document"`
	}
	filter := bson.M{"_id": objectID}
	opts := options.FindOne().SetProjection(bson.D{{Key: "document", Value: 1}})

	if err := repo.Collection.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewNotFoundError("CounterDocument", id))
		}
		errChan <- err
	}
	resultChan <- &result.Document
}

func (repo *counterRepository) Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error) {
	now := time.Now().UTC()
	counterDocument := NewCounterDocument(now)
	document := NewDocument(counterDocument, counterDocument.Id)
	event := NewCounterCreatedEvent()
	outbox := NewOutboxEvent(event, now)

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

func (repo *counterRepository) Patch(ctx context.Context, document *CounterDocument, patchModel *PatchModel, resultChan chan<- *CounterDocument, errChan chan<- error) {
	var result struct {
		Document CounterDocument `bson:"document"`
	}
	now := time.Now().UTC()
	event := NewCounterUpdatedEvent(patchModel.Increase, patchModel.UpdatedBy)
	outbox := NewOutboxEvent(event, now)

	filter := bson.D{{Key: "_id", Value: document.Id}, {Key: "document.version", Value: document.Version}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "document.counter", Value: document.Counter}}},
		{Key: "$inc", Value: bson.D{{Key: "document.version", Value: 1}}},
		{Key: "$set", Value: bson.D{{Key: "document.updatedAt", Value: now}}},
		{Key: "$set", Value: bson.D{{Key: "document.updatedBy", Value: document.UpdatedBy}}},
		{Key: "$addToSet", Value: bson.D{{Key: "outbox.events", Value: outbox}}},
	}
	options := options.FindOneAndUpdate().SetProjection(bson.D{{Key: "document", Value: 1}}).SetReturnDocument(options.After)

	if err := repo.Collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", document.Id)))
		}
		repo.Logger.Error("Could not update document", zap.Any("Document", document), zap.Error(err))
		errChan <- err
	}
	resultChan <- &result.Document
}
