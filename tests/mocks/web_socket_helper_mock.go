package mocks

import (
	"users-service/app/models/responses"

	"github.com/stretchr/testify/mock"
)

type MockWebSocketHelper struct {
	mock.Mock
}

func (m *MockWebSocketHelper) ForwardProgress(progressChan <-chan responses.Message) {
	m.Called(progressChan)
}

func (m *MockWebSocketHelper) SendMessage(msg responses.Message) {
	m.Called(msg)
}

func (m *MockWebSocketHelper) RespondFailure(message, errorDetail string) {
	m.Called(message, errorDetail)
}

func (m *MockWebSocketHelper) RespondSuccess(message, requestID string, data interface{}) {
	m.Called(message, requestID, data)
}

func (m *MockWebSocketHelper) ReadRequest(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}
