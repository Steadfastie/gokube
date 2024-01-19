package data

import "go.mongodb.org/mongo-driver/bson/primitive"

type DomainDocument interface{}

type Document struct {
	Id       primitive.ObjectID `bson:"_id"`
	Document DomainDocument     `bson:"document"`
	Outbox   OutboxBucket       `bson:"outbox"`
}

func NewDocument(data DomainDocument, id primitive.ObjectID) *Document {
	return &Document{
		Id:       id,
		Document: data,
		Outbox:   *NewOutboxBucket(),
	}
}
