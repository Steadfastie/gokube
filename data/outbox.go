package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType string

const (
	CounterCreated EventType = "Created"
	CounterUpdated EventType = "Updated"
)

type CounterCreatedEvent struct {
	Type EventType `bson:"type"`
}

func NewCounterCreatedEvent() *CounterCreatedEvent {
	return &CounterCreatedEvent{
		Type: CounterCreated,
	}
}

type CounterUpdatedEvent struct {
	Type      EventType `bson:"type"`
	Increased bool      `bson:"increased"`
	UpdatedBy string    `bson:"updatedBy"`
}

func NewCounterUpdatedEvent(increased bool, updatedBy string) *CounterUpdatedEvent {
	return &CounterUpdatedEvent{
		Type:      CounterUpdated,
		Increased: increased,
		UpdatedBy: updatedBy,
	}
}

type OutboxEvent struct {
	Id        primitive.ObjectID `bson:"_id"`
	Payload   any                `bson:"payload"`
	Timestamp time.Time          `bson:"timestamp"`
}

func NewOutboxEvent(event any, now time.Time) *OutboxEvent {
	return &OutboxEvent{
		Id:        primitive.NewObjectID(),
		Payload:   event,
		Timestamp: now,
	}
}

type OutboxBucket struct {
	LockId         *primitive.ObjectID `bson:"lockId, omitempty"`
	LockExpiration *time.Time          `bson:"lockExpiration, omitempty"`
	Events         []OutboxEvent       `bson:"events"`
}

func NewOutboxBucket() *OutboxBucket {
	return &OutboxBucket{
		LockId:         nil,
		LockExpiration: nil,
		Events:         []OutboxEvent{},
	}
}

func (outboxBucket *OutboxBucket) AddEvent(event *OutboxEvent) {
	outboxBucket.Events = append(outboxBucket.Events, *event)
}
