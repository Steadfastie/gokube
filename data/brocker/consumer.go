package brocker

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const groupId = "counter-consumer"

type Message struct {
	Type  string
	Value []byte
}

type Consumer interface {
	RecieveMessage(ctx context.Context, resultChan chan<- []byte, errChan chan<- error)
	CheckConnection() bool
	Disconnect()
}

type kafkaReader struct {
	Conn   *kafka.Conn
	Reader *kafka.Reader
	Logger *zap.Logger
}

func NewConsumer(ctx context.Context, logger *zap.Logger, addresses ...string) Consumer {
	conn, _ := kafka.DialLeader(ctx, "tcp", addresses[0], topic, 0)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   addresses,
		Topic:     topic,
		GroupID:   groupId,
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})

	connector := &kafkaReader{
		Conn:   conn,
		Reader: r,
		Logger: logger,
	}
	return connector
}

func (consumer *kafkaReader) Disconnect() {
	consumer.Reader.Close()
}

// Calls metadata endpoint. See https://github.com/segmentio/kafka-go/issues/389#issuecomment-569334516
func (consumer *kafkaReader) CheckConnection() bool {
	_, err := consumer.Conn.Brokers()
	return err == nil
}

func (consumer *kafkaReader) RecieveMessage(ctx context.Context, resultChan chan<- []byte, errChan chan<- error) {
	m, err := consumer.Reader.ReadMessage(ctx)
	if err != nil {
		consumer.Logger.Error(
			"Could not read a message from kafka",
			zap.String("key", string(m.Key)),
			zap.String("value", string(m.Value)),
			zap.String("topic", m.Topic),
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset),
			zap.Int64("messages in kafka", m.HighWaterMark),
		)
		errChan <- err
	}
	consumer.Logger.Info(
		"Recieved a message from kafka",
		zap.String("key", string(m.Key)),
		zap.String("topic", m.Topic),
		zap.Int("partition", m.Partition),
		zap.Int64("offset", m.Offset),
		zap.Int64("messages in kafka", m.HighWaterMark),
	)
	resultChan <- m.Value
}
