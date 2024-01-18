package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	id        primitive.ObjectID
	timestamp time.Time
	version   uint
}

type OutboxEvent struct {
	payload Event
}

type OutboxBucket struct {
	lockId         primitive.ObjectID `bson:"lockId"`
	lockExpiration time.Time          `bson:"lockExpiration"`
	events         []OutboxEvent
}
