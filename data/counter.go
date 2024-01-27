package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CounterDocument struct {
	Id        primitive.ObjectID `bson:"_id"`
	Counter   int
	Version   uint
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
	UpdatedBy string    `bson:"updatedBy,omitempty"`
}

func NewCounterDocument(now time.Time) *CounterDocument {
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

type CounterResponse struct {
	Id        primitive.ObjectID `example:"60c7c02ea38e3c3c4426c1bd"`
	Counter   int                `example:"5"`
	CreatedAt time.Time          `example:"2022-02-30T12:00:00Z"`
	UpdatedAt time.Time          `example:"2022-02-30T12:00:00Z"`
}

func (document *CounterDocument) MapToResponseModel() *CounterResponse {
	return &CounterResponse{
		Id:        document.Id,
		Counter:   document.Counter,
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
	}
}

type PatchCounterResponse struct {
	Before *CounterResponse `json:"before"`
	After  *CounterResponse `json:"after"`
}

func CreatePatchCounterResponse(before *CounterDocument, after *CounterDocument) *PatchCounterResponse {
	return &PatchCounterResponse{
		Before: before.MapToResponseModel(),
		After:  after.MapToResponseModel(),
	}
}
