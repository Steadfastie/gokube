package events

import (
	"time"

	"github.com/steadfastie/gokube/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CounterEvent struct {
	EventId primitive.ObjectID `bson:"_id" json:"id"`
	When    time.Time          `bson:"when" json:"when"`
	Who     string             `bson:"who" json:"who"`
	What    data.EventType     `bson:"what" json:"what"`
}
