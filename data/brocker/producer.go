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
	CheckConnection() bool
	Disconnect()
}

type kafkaWriter struct {
	Conn   *kafka.Conn
	Writer *kafka.Writer
	Logger *zap.Logger
}

func NewWriter(ctx context.Context, logger *zap.Logger, addresses ...string) Producer {
	conn, _ := kafka.DialLeader(ctx, "tcp", addresses[0], topic, 0)

	w := &kafka.Writer{
		Addr:                   kafka.TCP(addresses...),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           1,
		WriteTimeout:           10 * time.Second,
	}

	connector := &kafkaWriter{
		Conn:   conn,
		Writer: w,
		Logger: logger,
	}
	return connector
}

func (producer *kafkaWriter) Disconnect() {
	producer.Writer.Close()
}

// Calls metadata endpoint. See https://github.com/segmentio/kafka-go/issues/389#issuecomment-569334516
func (producer *kafkaWriter) CheckConnection() bool {
	_, err := producer.Conn.Brokers()
	return err == nil
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
