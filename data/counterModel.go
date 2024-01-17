package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CounterDocument struct {
	Id        primitive.ObjectID `bson:"_id" example:"60c7c02ea38e3c3c4426c1bd"`
	Counter   int                `example:"5"`
	Version   int                `example:"2"`
	CreatedAt time.Time          `example:"2022-02-30T12:00:00Z"`
	UpdatedAt time.Time          `example:"2022-02-30T12:00:00Z"`
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
