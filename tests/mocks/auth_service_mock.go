package mocks

import (
	"context"
	"users-service/app/models/ent"
	"users-service/app/models/requests"
	"users-service/app/models/responses"
	"users-service/app/models/utils"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateJWT(ctx context.Context, userID int, deviceID int, userRoles []utils.UserRoleInfo) (string, error) {
	args := m.Called(ctx, userID, deviceID, userRoles)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateJWT(ctx context.Context, tokenString string) (*utils.Claims, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.Claims), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CompareHashes(storedHash, password string) error {
	args := m.Called(storedHash, password)
	return args.Error(0)
}

func (m *MockAuthService) Register(ctx context.Context, req *requests.Register, reporter responses.Reporter) (*ent.User, error) {
	args := m.Called(ctx, req, reporter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *requests.Login, reporter responses.Reporter) (*ent.User, string, error) {
	args := m.Called(ctx, req, reporter)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.String(1), args.Error(2)
}

func (m *MockAuthService) ActivateAccount(ctx context.Context, req *requests.Code) (*ent.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockAuthService) VerifyAccount(ctx context.Context, req *requests.Code) (*ent.User, string, error) {
	args := m.Called(ctx, req)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.String(1), args.Error(2)
}

func (m *MockAuthService) VerifySecondFactor(ctx context.Context, req *requests.Code) (*ent.User, string, error) {
	args := m.Called(ctx, req)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.String(1), args.Error(2)
}

func (m *MockAuthService) RequestPasswordReset(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) CancelPasswordReset(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ConfirmPasswordReset(ctx context.Context, req *requests.ConfirmPasswordReset) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ResendActivationCode(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ResendVerificationCode(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ResendResetCode(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ResendSecondFactorCode(ctx context.Context, req *requests.Mail) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
