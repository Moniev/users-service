//go:build unit

package services

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"
	"users-service/app/models/ent"
	"users-service/app/models/requests"
	"users-service/tests/mocks"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(t *testing.T) (*AuthService, *mocks.MockUsersRepository, *mocks.MockEventNotifier) {
	mockRepo := new(mocks.MockUsersRepository)
	mockNotifier := new(mocks.MockEventNotifier)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	authService := NewAuthService(
		mockRepo,
		mockNotifier,
		zerolog.Nop(),
		pubKey,
		privKey,
		1,
	)
	require.NotNil(t, authService)

	return authService, mockRepo, mockNotifier
}

func TestAuthService_Register(t *testing.T) {
	req := &requests.Register{
		Mail:     "test@example.com",
		Password: "password123",
		Device: requests.Device{
			DeviceToken: "test-device",
		},
	}

	testCases := []struct {
		name          string
		setupMocks    func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectedUser  *ent.User
		expectErr     bool
		expectedError string
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, nil).Once()
				repo.On("CreateUser", mock.Anything, req, mock.AnythingOfType("string")).
					Return(&ent.User{ID: 1, Mail: req.Mail}, &ent.ActivationCode{Code: "123"}, nil).Once()
				notifier.On("CreateRegistrationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedUser: &ent.User{ID: 1, Mail: req.Mail},
			expectErr:    false,
		},
		{
			name: "Failure - User Already Exists",
			setupMocks: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(&ent.User{ID: 1}, nil).Once()
			},
			expectErr:     true,
			expectedError: "this email is already registered",
		},
		{
			name: "Failure - Repository Error on CreateUser",
			setupMocks: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, nil).Once()
				repo.On("CreateUser", mock.Anything, req, mock.AnythingOfType("string")).
					Return(nil, nil, errors.New("database error")).Once()
			},
			expectErr:     true,
			expectedError: "failed to create new user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMocks(mockRepo, mockNotifier)

			user, err := authService.Register(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUser.ID, user.ID)
			}

			mockRepo.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	authService, mockRepo, mockNotifier := newTestAuthService(t)
	authService.HashingPepper = []byte("test-pepper-for-login")

	password := "password123"
	hashedPassword, err := authService.HashPassword(password)
	require.NoError(t, err)

	req := &requests.Login{
		Mail:     "test@example.com",
		Password: password,
		Device:   requests.Device{DeviceToken: "test-device"},
	}

	userFromDB := &ent.User{
		ID:       1,
		Mail:     req.Mail,
		Password: hashedPassword,
		Active:   true,
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{TwoFactor: false},
		},
	}

	userWith2FA := &ent.User{
		ID:       2,
		Mail:     req.Mail,
		Password: hashedPassword,
		Active:   true,
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{TwoFactor: true},
		},
	}

	testCases := []struct {
		name           string
		setupMocks     func()
		expectErr      bool
		expectedToken  string
		expectedErrMsg string
	}{
		{
			name: "Success - No 2FA",
			setupMocks: func() {
				mockRepo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				mockRepo.On("FindOrCreateDevice", mock.Anything, userFromDB.ID, &req.Device).Return(&ent.UserDevice{ID: 1}, nil).Once()
				mockNotifier.On("CreateLoginEvent", mock.Anything, "password").Return(nil).Once()
			},
			expectErr:     false,
			expectedToken: "not-empty",
		},
		{
			name: "Success - 2FA Enabled",
			setupMocks: func() {
				mockRepo.On("GetUserByMail", mock.Anything, req.Mail).Return(userWith2FA, nil).Once()
				mockRepo.On("FindOrCreateDevice", mock.Anything, userWith2FA.ID, &req.Device).Return(&ent.UserDevice{ID: 2}, nil).Once()
				mockRepo.On("CreateSecondFactorCode", mock.Anything, userWith2FA, mock.Anything).Return(userWith2FA, &ent.SecondFactorCode{}, nil).Once()
				mockNotifier.On("CreateSecondFactorEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr:     false,
			expectedToken: "2fa_required",
		},
		{
			name: "Failure - User Not Found",
			setupMocks: func() {
				mockRepo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("not found")).Once()
			},
			expectErr:      true,
			expectedErrMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.Mock = mock.Mock{}
			mockNotifier.Mock = mock.Mock{}
			tc.setupMocks()

			_, token, err := authService.Login(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				if tc.expectedToken == "not-empty" {
					assert.NotEmpty(t, token)
				} else {
					assert.Equal(t, tc.expectedToken, token)
				}
			}

			mockRepo.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}

func TestAuthService_ActivateAccount(t *testing.T) {
	req := &requests.Code{Code: "valid-code"}
	userFromDB := &ent.User{ID: 1, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("ActivateAccount", mock.Anything, req.Code).Return(userFromDB, nil).Once()
				notifier.On("CreateUserActionEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Invalid Code",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("ActivateAccount", mock.Anything, req.Code).Return(nil, errors.New("invalid code")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			_, err := authService.ActivateAccount(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_VerifyAccount(t *testing.T) {
	req := &requests.Code{
		Code: "valid-code",
		Device: requests.Device{
			DeviceToken: "verify-device",
		},
	}
	userFromDB := &ent.User{
		ID:    1,
		Phone: "+123456789",
		Edges: ent.UserEdges{
			UserSettings:     &ent.UserSettings{},
			VerificationCode: &ent.VerificationCode{},
		},
	}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("VerifyAccount", mock.Anything, req.Code).Return(userFromDB, nil).Once()
				repo.On("FindOrCreateDevice", mock.Anything, userFromDB.ID, &req.Device).Return(&ent.UserDevice{ID: 2}, nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Invalid Verification Code",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("VerifyAccount", mock.Anything, req.Code).Return(nil, errors.New("invalid code")).Once()
			},
			expectErr: true,
		},
		{
			name: "Failure - Device Creation Fails",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("VerifyAccount", mock.Anything, req.Code).Return(userFromDB, nil).Once()
				repo.On("FindOrCreateDevice", mock.Anything, userFromDB.ID, &req.Device).Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			_, token, err := authService.VerifyAccount(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
			mockRepo.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}

func TestAuthService_VerifySecondFactor(t *testing.T) {
	req := &requests.Code{Code: "123456"}
	device := &ent.UserDevice{ID: 1}
	userFromDB := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{},
			SecondFactorCode: &ent.SecondFactorCode{
				ExpiresAt: time.Now().Add(time.Minute * 5),
				Edges: ent.SecondFactorCodeEdges{
					TargetUserDevice: device,
				},
			},
		},
	}
	userWithExpiredCode := &ent.User{
		ID: 2,
		Edges: ent.UserEdges{
			SecondFactorCode: &ent.SecondFactorCode{
				ExpiresAt: time.Now().Add(-time.Minute * 5),
			},
		},
	}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserBySecondFactor", mock.Anything, req.Code).Return(userFromDB, nil).Once()
				repo.On("RemoveSecondFactorCode", mock.Anything, userFromDB).Return(userFromDB, nil).Once()
				notifier.On("CreateLoginEvent", mock.Anything, "second-factor").Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Invalid Code",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserBySecondFactor", mock.Anything, req.Code).Return(nil, errors.New("code not found")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user by second factor code",
		},
		{
			name: "Failure - Expired Code",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserBySecondFactor", mock.Anything, req.Code).Return(userWithExpiredCode, nil).Once()
				repo.On("RemoveSecondFactorCode", mock.Anything, userWithExpiredCode).Return(userWithExpiredCode, nil).Maybe()
			},
			expectErr: true,
			errMsg:    "token has expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			_, token, err := authService.VerifySecondFactor(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RequestPasswordReset(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userFromDB := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{},
		},
	}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				repo.On("CreateResetCode", mock.Anything, userFromDB).Return(&ent.ResetCode{}, nil).Once()
				notifier.On("CreateResetPasswordEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("user not found")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user",
		},
		{
			name: "Failure - Create Code Fails",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				repo.On("CreateResetCode", mock.Anything, userFromDB).Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to create reset token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.RequestPasswordReset(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}

func TestAuthService_ConfirmPasswordReset(t *testing.T) {
	authService, mockRepo, mockNotifier := newTestAuthService(t)
	authService.HashingPepper = []byte("test-pepper")

	req := &requests.ConfirmPasswordReset{
		Code:     "valid-reset-code",
		Password: "NewStrongPassword123!",
	}
	userFromDB := &ent.User{ID: 1, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func()
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			setupMock: func() {
				mockRepo.On("GetUserByResetCode", mock.Anything, req.Code).Return(userFromDB, nil).Once()
				mockRepo.On("UpdateUsersPassword", mock.Anything, userFromDB, mock.AnythingOfType("string")).Return(userFromDB, nil).Once()
				mockRepo.On("RemoveResetCode", mock.Anything, userFromDB).Return(userFromDB, nil).Once()
				mockNotifier.On("CreateUserActionEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Invalid Reset Code",
			setupMock: func() {
				mockRepo.On("GetUserByResetCode", mock.Anything, req.Code).Return(nil, errors.New("failed to fetch user")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.Mock = mock.Mock{}
			mockNotifier.Mock = mock.Mock{}
			tc.setupMock()

			err := authService.ConfirmPasswordReset(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_CancelPasswordReset(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userFromDB := &ent.User{ID: 1, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				repo.On("RemoveResetCode", mock.Anything, userFromDB).Return(userFromDB, nil).Once()
				notifier.On("CreateUserActionEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("failed to fetch user")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.CancelPasswordReset(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ResendActivationCode(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userFromDB := &ent.User{ID: 1, Edges: ent.UserEdges{ActivationCode: &ent.ActivationCode{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				notifier.On("CreateRegistrationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.ResendActivationCode(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ResendVerificationCode(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userFromDB := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			UserSettings:     &ent.UserSettings{},
			VerificationCode: &ent.VerificationCode{},
		},
	}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				notifier.On("CreateVerificationEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.ResendVerificationCode(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ResendResetCode(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userFromDB := &ent.User{ID: 1, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userFromDB, nil).Once()
				repo.On("CreateResetCode", mock.Anything, userFromDB).Return(&ent.ResetCode{}, nil).Once()
				notifier.On("CreateResetPasswordEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.ResendResetCode(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ResendSecondFactorCode(t *testing.T) {
	req := &requests.Mail{Mail: "user@example.com"}
	userWith2FA := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			UserSettings:     &ent.UserSettings{TwoFactor: true},
			SecondFactorCode: &ent.SecondFactorCode{},
		},
	}
	userWithout2FA := &ent.User{
		ID: 2,
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{TwoFactor: false},
		},
	}

	testCases := []struct {
		name      string
		user      *ent.User
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			user: userWith2FA,
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userWith2FA, nil).Once()
				notifier.On("CreateSecondFactorEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - 2FA Not Enabled",
			user: userWithout2FA,
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByMail", mock.Anything, req.Mail).Return(userWithout2FA, nil).Once()
			},
			expectErr: true,
			errMsg:    "second factor is not enabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockNotifier := newTestAuthService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := authService.ResendSecondFactorCode(context.Background(), req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateJWT(t *testing.T) {
	authService, _, _ := newTestAuthService(t)

	authService.TokenDuration = 15 * time.Minute

	validToken, err := authService.GenerateJWT(context.Background(), 1, 1, nil)
	require.NoError(t, err)

	authService.TokenDuration = -15 * time.Minute
	expiredToken, err := authService.GenerateJWT(context.Background(), 1, 1, nil)
	require.NoError(t, err)
	authService.TokenDuration = 15 * time.Minute

	invalidAuthService, _, _ := newTestAuthService(t)
	invalidToken, err := invalidAuthService.GenerateJWT(context.Background(), 1, 1, nil)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		token     string
		service   *AuthService
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Success - Valid Token",
			token:     validToken,
			service:   authService,
			expectErr: false,
		},
		{
			name:      "Failure - Expired Token",
			token:     expiredToken,
			service:   authService,
			expectErr: true,
			errMsg:    "token has invalid claims: token is expired",
		},
		{
			name:      "Failure - Invalid Signature",
			token:     invalidToken,
			service:   authService,
			expectErr: true,
			errMsg:    "token signature is invalid: ed25519: verification error",
		},
		{
			name:      "Failure - Malformed Token",
			token:     "this.is.not.a.jwt",
			service:   authService,
			expectErr: true,
			errMsg:    "token is malformed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := tc.service.ValidateJWT(context.Background(), tc.token)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, 1, claims.UserID)
			}
		})
	}
}
