package brocker

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Producer interface {
	SendMessage(topic string, value []byte)
}

type kafkaConnector struct {
	Producer *kafka.Producer
	Logger   *zap.Logger
}

func NewConnector(bootstrapServer string, logger *zap.Logger) (Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServer})
	if err != nil {
		return nil, fmt.Errorf("Kafka producer could not be established: %w", err)
	}

	go func(producer *kafka.Producer, logger *zap.Logger) {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			case kafka.Error:
				logger.Error("Kafka error", zap.Error(ev))
			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}(p, logger)

	connector := &kafkaConnector{
		Producer: p,
		Logger:   logger,
	}
	return connector, nil
}

func (connector *kafkaConnector) SendMessage(topic string, value []byte) {
	connector.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          value,
		Headers:        []kafka.Header{{Key: "eventType", Value: []byte("header values are binary")}},
	}, nil)
}
