package handlers

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConsumerInterface interface {
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error)
	Stop()
}

type ConsumerWrapper struct {
	Consumer ConsumerInterface
	Cancel   context.CancelFunc
}

func NewConsumerWrapper(consumer ConsumerInterface, cancel context.CancelFunc) *ConsumerWrapper {
	return &ConsumerWrapper{Consumer: consumer, Cancel: cancel}
}

func (w *ConsumerWrapper) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	return w.Consumer.ReadMessage(timeout)
}

func (w *ConsumerWrapper) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	return w.Consumer.GetMetadata(topic, allTopics, timeoutMs)
}

func (w *ConsumerWrapper) CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error) {
	return w.Consumer.CommitMessage(msg)
}

func (w *ConsumerWrapper) Stop() {
	w.Cancel()
}

type MessageHandler interface {
	Handle(ctx context.Context, msg *kafka.Message) error
}

type KafkaConsumerAdapter struct {
	Consumer *kafka.Consumer
	Cancel   context.CancelFunc
}

var _ ConsumerInterface = (*KafkaConsumerAdapter)(nil)

func (a *KafkaConsumerAdapter) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	return a.Consumer.ReadMessage(timeout)
}

func (a *KafkaConsumerAdapter) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	return a.Consumer.GetMetadata(topic, allTopics, timeoutMs)
}

func (a *KafkaConsumerAdapter) CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error) {
	return a.Consumer.CommitMessage(msg)
}

func (a *KafkaConsumerAdapter) Stop() {
	_ = a.Consumer.Close()
}
