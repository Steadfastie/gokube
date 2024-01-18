package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CounterDocument struct {
	Id        primitive.ObjectID `bson:"_id" example:"60c7c02ea38e3c3c4426c1bd"`
	Counter   int                `example:"5"`
	Version   int                `json:"-" example:"2"`
	CreatedAt time.Time          `bson:"createdAt" example:"2022-02-30T12:00:00Z"`
	UpdatedAt time.Time          `bson:"updatedAt" example:"2022-02-30T12:00:00Z"`
	UpdatedBy string             `bson:"updatedBy,omitempty" json:"updatedBy,omitempty" example:"user"`
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

func (document *CounterDocument) Copy() *CounterDocument {
	return &CounterDocument{
		Id:        document.Id,
		Counter:   document.Counter,
		Version:   document.Version,
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
		UpdatedBy: document.UpdatedBy,
	}
}

func (document *CounterDocument) IncreaseCounter(updatedBy string) {
	document.Counter++
	document.UpdatedAt = time.Now().UTC()
	document.UpdatedBy = updatedBy
}

func (document *CounterDocument) DecreaseCounter(updatedBy string) {
	document.Counter--
	document.UpdatedAt = time.Now().UTC()
	document.UpdatedBy = updatedBy
}
