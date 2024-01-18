package data

type DomainDocument interface{}

type Document struct {
	document DomainDocument
	outbox   OutboxBucket
}

func NewDocument(data DomainDocument) *Document {
	return &Document{
		document: data,
		outbox:   OutboxBucket{},
	}
}
