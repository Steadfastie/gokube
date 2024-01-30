package brocker

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/steadfastie/gokube/data"
	"go.uber.org/zap"
)

const topic = "counter"

type Producer interface {
	SendMessage(ctx context.Context, key []byte, value []byte)
	Disconnect()
}

type kafkaWriter struct {
	Writer *kafka.Writer
	Logger *zap.Logger
}

func NewWriter(ctx context.Context, logger *zap.Logger, addresses ...string) Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(addresses...),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           1,
		WriteTimeout:           10 * time.Second,
	}

	connector := &kafkaWriter{
		Writer: w,
		Logger: logger,
	}
	return connector
}

func (producer *kafkaWriter) Disconnect() {
	producer.Writer.Close()
}

func (producer *kafkaWriter) SendMessage(ctx context.Context, key []byte, value []byte) {
	retryConfig := &data.RetryConfig{
		Context:           ctx,
		Logger:            producer.Logger,
		RecoverableErrors: []error{kafka.LeaderNotAvailable, context.DeadlineExceeded},
	}

	err := data.WithRetry(retryConfig, func() error {
		return producer.Writer.WriteMessages(retryConfig.Context,
			kafka.Message{
				Key:   key,
				Value: value,
			},
		)
	})
	if err != nil {
		producer.Logger.Error("Could not write a message to kafka", zap.Error(err))
	}
}
