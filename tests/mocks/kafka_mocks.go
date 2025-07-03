package mocks

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

type MockConsumerWrapper struct {
	mock.Mock
}

func (m *MockConsumerWrapper) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	args := m.Called(timeout)
	var msg *kafka.Message
	if args.Get(0) != nil {
		msg = args.Get(0).(*kafka.Message)
	}
	return msg, args.Error(1)
}

func (m *MockConsumerWrapper) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	args := m.Called(topic, allTopics, timeoutMs)
	var metadata *kafka.Metadata
	if args.Get(0) != nil {
		metadata = args.Get(0).(*kafka.Metadata)
	}
	return metadata, args.Error(1)
}

func (m *MockConsumerWrapper) CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error) {
	args := m.Called(msg)
	var tps []kafka.TopicPartition
	if args.Get(0) != nil {
		tps = args.Get(0).([]kafka.TopicPartition)
	}
	return tps, args.Error(1)
}
