package infrastucture

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"time"
)

type MongoDB struct {
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
}

func NewMongoClient(ctx context.Context, config *Config, logger *zap.Logger) (*MongoDB, error) {
	logger.Debug("Trying to connect to MongoDB")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+config.MongoSettings.MongoConnectionString))
	if err != nil {
		logger.Error("Could not connect to MongoDB", zap.Error(err))
		panic(err)
	}

	ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error("Connection to MongoDB established unsuccessfully", zap.Error(err))
		panic(err)
	}

	mongodb := &MongoDB{
		MongoClient: client,
		MongoDB:     client.Database(config.MongoSettings.MongoDatabase),
	}

	logger.Debug("Connection to MongoDB is ready")
	return mongodb, err
}

func (mongodb *MongoDB) DisconnectMongoClient(ctx context.Context, logger *zap.Logger) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := mongodb.MongoClient.Disconnect(ctx); err != nil {
		logger.Panic("Could not disconnect mongo client", zap.Error(err))
	}
}
