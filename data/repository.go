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

const Collection = "basic"

type CounterRepository interface {
	GetById(ctx context.Context, id string, resultChan chan<- *CounterDocument, errChan chan<- error)
	Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error)
	Patch(ctx context.Context, document *CounterDocument, resultChan chan<- *CounterDocument, errChan chan<- error)
}

type counterRepository struct {
	Collection *mongo.Collection
	Logger     *zap.Logger
}

func NewCounterRepository(mongodb *services.MongoDB, logger *zap.Logger) CounterRepository {
	return &counterRepository{
		Collection: mongodb.MongoDB.Collection(Collection),
		Logger:     logger,
	}
}

func (repo *counterRepository) GetById(ctx context.Context, id string, resultChan chan<- *CounterDocument, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	var result CounterDocument
	filter := bson.M{"_id": objectID}

	if err := repo.Collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewNotFoundError("CounterDocument", id))
		}
		errChan <- err
	}
	resultChan <- &result
}

func (repo *counterRepository) Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error) {
	document := NewCounterDocument()

	result, err := repo.Collection.InsertOne(ctx, document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", document.Id)))
		}
		repo.Logger.Error("Could not create document", zap.Any("Document", document), zap.Error(err))
		errChan <- err
	} else {
		resultChan <- result.InsertedID.(primitive.ObjectID)
	}
}

func (repo *counterRepository) Patch(ctx context.Context, document *CounterDocument, resultChan chan<- *CounterDocument, errChan chan<- error) {
	var result CounterDocument

	filter := bson.D{{Key: "_id", Value: document.Id}, {Key: "version", Value: document.Version}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "counter", Value: document.Counter}}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
		{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: time.Now().UTC()}}},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	if err := repo.Collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", document.Id)))
		}
		repo.Logger.Error("Could not update document", zap.Any("Document", document), zap.Error(err))
		errChan <- err
	} else {
		resultChan <- &result
	}
}
