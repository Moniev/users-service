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

func (m *MockUsersService) UpdateEntrepreneurDetails(
	ctx context.Context,
	userID int,
	req *requests.EntrepreneurDetails,
) (*ent.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) UpdateLocation(
	ctx context.Context,
	userID int,
	req *requests.Location,
) (*ent.User, error) {

	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) GetUserPrivate(
	ctx context.Context,
	userID int,
) (*ent.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) AddRole(ctx context.Context, userID, roleID int) (*ent.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) RevokeRole(ctx context.Context, userID, roleID int) (*ent.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockUsersService) Index(ctx context.Context, page, pageSize int) ([]*ent.User, error) {
	args := m.Called(ctx, page, pageSize)

	var r0 []*ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*ent.User)
	}

	return r0, args.Get(1).(error)
}
