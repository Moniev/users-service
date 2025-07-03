package mocks

import (
	"context"
	"users-service/app/models/events"

	"github.com/stretchr/testify/mock"
)

type MockEventListener struct {
	mock.Mock
}

func (m *MockEventListener) HandleUserActionEvent(ctx context.Context, event *events.UserActionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventListener) Listen(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventListener) Ping() error {
	args := m.Called()
	return args.Error(0)
}
