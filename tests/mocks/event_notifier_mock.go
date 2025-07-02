package mocks

import (
	"users-service/app/models/ent"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	args := m.Called(msg, deliveryChan)
	return args.Error(0)
}

type MockEventNotifier struct {
	mock.Mock
}

func (m *MockEventNotifier) CreateRegistrationEvent(user *ent.User, activationCode *ent.ActivationCode) error {
	args := m.Called(user, activationCode)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateLoginEvent(settings *ent.UserSettings, loginMethod string) error {
	args := m.Called(settings, loginMethod)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateSecondFactorEvent(settings *ent.UserSettings, secondFactor *ent.SecondFactorCode) error {
	args := m.Called(settings, secondFactor)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateVerificationEvent(settings *ent.UserSettings, phone string, verificationCode *ent.VerificationCode) error {
	args := m.Called(settings, phone, verificationCode)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateResetPasswordEvent(settings *ent.UserSettings, resetCode *ent.ResetCode) error {
	args := m.Called(settings, resetCode)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateNotificationEvent(settings *ent.UserSettings, devices []*ent.UserDevice) error {
	args := m.Called(settings, devices)
	return args.Error(0)
}

func (m *MockEventNotifier) CreateUserActionEvent(settings *ent.UserSettings, action *ent.UserAction) error {
	args := m.Called(settings, action)
	return args.Error(0)
}

func (m *MockEventNotifier) Produce(key []byte, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}
