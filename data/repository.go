package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/infrastucture"
	domainErrors "github.com/steadfastie/gokube/infrastucture/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const Collection = "basic"

type BasicDocument struct {
	id        primitive.ObjectID `bson:"_id"`
	counter   int
	version   int
	createdAt time.Time
	updatedAt time.Time
}

func GetById(ctx context.Context, id string, resultChan chan<- *BasicDocument, errChan chan<- error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errChan <- err
		return
	}

	go func() {
		var result BasicDocument
		filter := bson.M{"_id": objectID}

		container.Call(func(logger *zap.Logger, mongodb *infrastucture.MongoDB) error {
			if err := mongodb.MongoDB.Collection(Collection).FindOne(ctx, filter).Decode(&result); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					logger.Error("Could not find document with id", zap.Any("id", objectID))
				}
				resultChan <- nil
				errChan <- err
			}
			resultChan <- &result
			return err
		})
	}()
}

func Create(ctx context.Context, resultChan chan<- *primitive.ObjectID, errChan chan<- error) {
	now := time.Now().UTC()
	document := &BasicDocument{
		id:        primitive.NewObjectID(),
		counter:   0,
		version:   0,
		createdAt: now,
		updatedAt: now,
	}

	go func() {
		container.Call(func(logger *zap.Logger, mongodb *infrastucture.MongoDB) {
			_, err := mongodb.MongoDB.Collection(Collection).InsertOne(ctx, document)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", document.id)))
				}
				logger.Error("Could not create document", zap.Any("Document", document), zap.Error(err))
				errChan <- err
			} else {
				resultChan <- &document.id
			}
		})
	}()
}

func Patch(ctx context.Context, document *BasicDocument, resultChan chan<- *BasicDocument, errChan chan<- error) {
	go func() {
		var result BasicDocument
		filter := bson.D{{Key: "_id", Value: document.id}, {Key: "version", Value: document.version}}
		update := bson.D{
			{Key: "$set", Value: bson.D{{Key: "counter", Value: document.counter}}},
			{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
			{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: time.Now().UTC()}}},
		}
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)
		container.Call(func(logger *zap.Logger, mongodb *infrastucture.MongoDB) {
			if err := mongodb.MongoDB.Collection(Collection).FindOneAndUpdate(ctx, filter, update, options).Decode(&result); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					panic(domainErrors.NewOptimisticLockError(fmt.Sprintf("Document with id - {%v} - has already been modified", document.id)))
				}
				logger.Error("Could not update document", zap.Any("Document", document), zap.Error(err))
				errChan <- err
			} else {
				resultChan <- &result
			}
		})
	}()
}
