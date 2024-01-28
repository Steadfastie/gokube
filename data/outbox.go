package data

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType string

const (
	CounterCreated EventType = "Created"
	CounterUpdated EventType = "Updated"
)

type CounterCreatedEvent struct {
	Type      EventType `bson:"type"`
	UserAlias string    `bson:"userAlias"`
}

func NewCounterCreatedEvent(userAlias string) *CounterCreatedEvent {
	return &CounterCreatedEvent{
		Type:      CounterCreated,
		UserAlias: userAlias,
	}
}

type CounterUpdatedEvent struct {
	Type      EventType `bson:"type"`
	Counter   int       `bson:"counter"`
	UpdatedBy string    `bson:"updatedBy"`
	UserAlias string    `bson:"userAlias"`
}

func NewCounterUpdatedEvent(counter int, updatedBy string, userAlias string) *CounterUpdatedEvent {
	return &CounterUpdatedEvent{
		Type:      CounterUpdated,
		Counter:   counter,
		UpdatedBy: updatedBy,
		UserAlias: userAlias,
	}
}

type EventPayload interface{}

type OutboxEvent struct {
	Id        primitive.ObjectID `bson:"_id"`
	Payload   any                `bson:"payload"`
	Timestamp time.Time          `bson:"timestamp"`
}

func (event *OutboxEvent) UnmarshalBSON(data []byte) error {
	raw := bson.Raw(data)

	id, ok := raw.Lookup("_id").ObjectIDOK()
	if !ok {
		return errors.New(`failed to find field "_id"`)
	}
	event.Id = id

	timestamp, ok := raw.Lookup("timestamp").TimeOK()
	if !ok {
		return errors.New(`failed to find field "timestamp"`)
	}
	event.Timestamp = timestamp

	payload := raw.Lookup("payload").Document()
	payloadType, ok := payload.Lookup("type").StringValueOK()
	if !ok {
		return errors.New(`payload did not contain field "type"`)
	}

	switch payloadType {
	case string(CounterCreated):
		event.Payload = &CounterCreatedEvent{
			Type:      CounterUpdated,
			UserAlias: payload.Lookup("userAlias").StringValue(),
		}
	case string(CounterUpdated):
		event.Payload = &CounterUpdatedEvent{
			Type:      CounterUpdated,
			Counter:   int(payload.Lookup("counter").AsInt64()),
			UpdatedBy: payload.Lookup("updatedBy").StringValue(),
			UserAlias: payload.Lookup("userAlias").StringValue(),
		}
	default:
		return fmt.Errorf("unknown event type %q", payloadType)
	}

	return nil
}

type ByTimestamp []OutboxEvent

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

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
