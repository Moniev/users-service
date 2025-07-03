package utils

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConsumerWrapper struct {
	Consumer ConsumerWrapperInterface
	Cancel   context.CancelFunc
}

type ConsumerWrapperInterface interface {
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error)
}

func NewConsumerWrapper(consumer ConsumerWrapperInterface, cancel context.CancelFunc) *ConsumerWrapper { // Przykład
	return &ConsumerWrapper{Consumer: consumer, Cancel: cancel}
}
