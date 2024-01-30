package events

import (
	"time"

	"github.com/steadfastie/gokube/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServiceName string

const (
	Api      ServiceName = "api"
	Outbox   ServiceName = "outbox"
	Consumer ServiceName = "consumer"
)

type CounterEvent struct {
	EventId   primitive.ObjectID `bson:"_id" json:"id"`
	CounterId primitive.ObjectID `bson:"counterId" json:"counterId"`
	Who       string             `bson:"who" json:"who"`
	What      data.EventType     `bson:"what" json:"what"`
	Trail     []Trail            `bson:"trail" json:"trail"`
}

type Trail struct {
	Service   ServiceName `bson:"service" json:"service"`
	Timestamp time.Time   `bson:"timestamp" json:"timestamp"`
}

func (event *CounterEvent) AddTrail(service ServiceName, timestamp time.Time) {
	event.Trail = append(event.Trail, Trail{
		Service:   service,
		Timestamp: timestamp,
	})
}
