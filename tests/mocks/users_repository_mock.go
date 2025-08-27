package mocks

import (
	"context"
	"users-service/app/models/ent"
	"users-service/app/models/requests"

	"github.com/stretchr/testify/mock"
)

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) CreateUser(ctx context.Context, req *requests.Register, hashedPassword string) (*ent.User, *ent.ActivationCode, error) {
	args := m.Called(ctx, req, hashedPassword)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	var r1 *ent.ActivationCode
	if args.Get(1) != nil {
		r1 = args.Get(1).(*ent.ActivationCode)
	}
	return r0, r1, args.Error(2)
}

func (m *MockUsersRepository) CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error) {
	args := m.Called(ctx, user)
	var r0 *ent.ResetCode
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.ResetCode)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) CreateSecondFactorCode(ctx context.Context, user *ent.User, device *ent.UserDevice) (*ent.SecondFactorCode, error) {
	args := m.Called(ctx, user, device)

	var r0 *ent.SecondFactorCode
	if args.Get(1) != nil {
		r0 = args.Get(1).(*ent.SecondFactorCode)
	}

	return r0, args.Error(2)
}

func (m *MockUsersRepository) GetUserByID(ctx context.Context, ID int) (*ent.User, error) {
	args := m.Called(ctx, ID)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserByMail(ctx context.Context, mail string) (*ent.User, error) {
	args := m.Called(ctx, mail)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserByPhone(ctx context.Context, phone string) (*ent.User, error) {
	args := m.Called(ctx, phone)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error) {
	args := m.Called(ctx, code)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error) {
	args := m.Called(ctx, code)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserByResetCode(ctx context.Context, code string) (*ent.User, error) {
	args := m.Called(ctx, code)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) FindOrCreateDevice(ctx context.Context, userID int, req *requests.Device) (*ent.UserDevice, error) {
	args := m.Called(ctx, userID, req)
	var r0 *ent.UserDevice
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.UserDevice)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) ActivateAccount(ctx context.Context, code string) (*ent.User, error) {
	args := m.Called(ctx, code)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) VerifyAccount(ctx context.Context, code string) (*ent.User, error) {
	args := m.Called(ctx, code)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user *ent.User, req *requests.User) (*ent.User, error) {
	args := m.Called(ctx, user, req)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error) {
	args := m.Called(ctx, user, hashedPassword)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) UpdateUsersDetails(ctx context.Context, user *ent.User, req *requests.Details) (*ent.User, error) {
	args := m.Called(ctx, user, req)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) UpdateUsersSettings(ctx context.Context, user *ent.User, req *requests.Settings) (*ent.User, error) {
	args := m.Called(ctx, user, req)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	args := m.Called(ctx, user)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}
	return r0, args.Error(1)
}

func (m *MockUsersRepository) RemoveSecondFactorCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	args := m.Called(ctx, user)
	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) RemoveAccount(ctx context.Context, user *ent.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUsersRepository) UpdateEntrepreneurDetails(
	ctx context.Context,
	user *ent.User,
	req *requests.EntrepreneurDetails,
) (*ent.User, error) {
	args := m.Called(ctx, user, req)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) UpdateLocation(
	ctx context.Context,
	user *ent.User,
	req *requests.Location,
) (*ent.User, error) {
	args := m.Called(ctx, user, req)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) AddSubscriptions(
	ctx context.Context,
	user *ent.User,
	subIDs []int,
) error {
	args := m.Called(ctx, user, subIDs)
	return args.Error(0)
}

func (m *MockUsersRepository) RemoveSubscriptions(
	ctx context.Context,
	user *ent.User,
	subIDs []int,
) error {
	args := m.Called(ctx, user, subIDs)
	return args.Error(0)
}

func (m *MockUsersRepository) CreateUserAction(
	ctx context.Context,
	user *ent.User,
	action, originDevice, details string,
) error {

	args := m.Called(ctx, user, action, originDevice, details)
	return args.Error(0)
}

func (m *MockUsersRepository) AddRole(
	ctx context.Context,
	userID, roleID int,
) (*ent.User, error) {
	args := m.Called(ctx, userID, roleID)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) RevokeRole(
	ctx context.Context,
	userID, roleID int,
) (*ent.User, error) {
	args := m.Called(ctx, userID, roleID)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserByMailWithCodes(ctx context.Context, mail string) (*ent.User, error) {
	args := m.Called(ctx, mail)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserFunctionalByID(ctx context.Context, ID int) (*ent.User, error) {
	args := m.Called(ctx, ID)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserFunctionalDetailsByID(ctx context.Context, ID int) (*ent.User, error) {
	args := m.Called(ctx, ID)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUsersPublic(ctx context.Context, page, pageSize int) ([]*ent.User, error) {
	args := m.Called(ctx, page, pageSize)

	var r0 []*ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*ent.User)
	}

	return r0, args.Error(1)
}

func (m *MockUsersRepository) GetUserPublicByID(ctx context.Context, ID int) (*ent.User, error) {
	args := m.Called(ctx, ID)

	var r0 *ent.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*ent.User)
	}

	return r0, args.Error(1)
}
