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

type BasicRepository interface {
	GetById(ctx context.Context, id string, resultChan chan<- *BasicDocument, errChan chan<- error)
	Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error)
	Patch(ctx context.Context, document *BasicDocument, resultChan chan<- *BasicDocument, errChan chan<- error)
}

type basicRepository struct {
	Collection *mongo.Collection
	Logger     *zap.Logger
}

func NewBasicRepository(mongodb *services.MongoDB, logger *zap.Logger) BasicRepository {
	return &basicRepository{
		Collection: mongodb.MongoDB.Collection(Collection),
		Logger:     logger,
	}
}

type BasicDocument struct {
	Id        primitive.ObjectID `bson:"_id"`
	Counter   int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBasicDocument() *BasicDocument {
	now := time.Now().UTC()
	return &BasicDocument{
		Id:        primitive.NewObjectID(),
		Counter:   0,
		Version:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (repo *basicRepository) GetById(ctx context.Context, id string, resultChan chan<- *BasicDocument, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	var result BasicDocument
	filter := bson.M{"_id": objectID}

	if err := repo.Collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			repo.Logger.Error("Could not find document with id", zap.Any("id", objectID))
		}
		errChan <- err
	}
	resultChan <- &result
}

func (repo *basicRepository) Create(ctx context.Context, resultChan chan<- primitive.ObjectID, errChan chan<- error) {
	document := NewBasicDocument()

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

func (repo *basicRepository) Patch(ctx context.Context, document *BasicDocument, resultChan chan<- *BasicDocument, errChan chan<- error) {
	var result BasicDocument

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
