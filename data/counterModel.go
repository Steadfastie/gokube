package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CounterDocument struct {
	Id        primitive.ObjectID `bson:"_id"`
	Counter   int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCounterDocument() *CounterDocument {
	now := time.Now().UTC()
	return &CounterDocument{
		Id:        primitive.NewObjectID(),
		Counter:   0,
		Version:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
