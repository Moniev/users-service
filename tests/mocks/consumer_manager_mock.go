package mocks

import (
	"users-service/app/models/handlers"

	"github.com/stretchr/testify/mock"
)

type MockConsumerManager struct {
	mock.Mock
}

func (m *MockConsumerManager) Register(topic string, consumer handlers.ConsumerInterface) error {
	args := m.Called(topic, consumer)
	return args.Error(0)
}

func (m *MockConsumerManager) PingAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConsumerManager) GetConsumer(topic string) (handlers.ConsumerInterface, bool) {
	args := m.Called(topic)
	return args.Get(0).(handlers.ConsumerInterface), args.Bool(1)
}
