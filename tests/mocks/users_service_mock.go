// tests/mocks/users_service_mock.go
package mocks

import (
	"context"
	"users-service/app/models/ent"
	"users-service/app/models/requests"

	"github.com/stretchr/testify/mock"
)

type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) UpdateUser(ctx context.Context, userID int, req *requests.User) (*ent.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) UpdateDetails(ctx context.Context, userID int, req *requests.Details) (*ent.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) UpdateSettings(ctx context.Context, userID int, req *requests.Settings) (*ent.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) RemoveAccount(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUsersService) GetUserPublic(ctx context.Context, userID int) (*ent.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}
