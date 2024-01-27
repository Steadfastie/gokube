package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type MongoDB struct {
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
}

type MongoSettings struct {
	ConnectionString string `json:"MongoConnectionString"`
	Database         string `json:"MongoDatabase"`
}

type MongoSettingsProvider interface {
	GetMongoSettings() MongoSettings
}

func NewMongoClient(ctx context.Context, config MongoSettingsProvider, logger *zap.Logger) (*MongoDB, error) {
	mongoSettings := config.GetMongoSettings()
	logger.Debug("Trying to connect to MongoDB")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+mongoSettings.ConnectionString))
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
		MongoDB:     client.Database(mongoSettings.Database),
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
