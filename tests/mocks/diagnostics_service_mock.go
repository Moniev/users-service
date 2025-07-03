package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockDiagnosticsService struct {
	mock.Mock
}

func (m *MockDiagnosticsService) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDiagnosticsService) CheckReadiness(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
