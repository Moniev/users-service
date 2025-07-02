//go:build unit

package services

import (
	"context"
	"errors"
	"testing"
	"users-service/app/models/ent"
	"users-service/app/models/requests"
	"users-service/tests/mocks"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestUsersService(t *testing.T) (*UsersService, *mocks.MockUsersRepository, *mocks.MockEventNotifier) {
	mockRepo := new(mocks.MockUsersRepository)
	mockNotifier := new(mocks.MockEventNotifier)

	usersService := NewUsersService(mockRepo, mockNotifier, zerolog.Nop())
	require.NotNil(t, usersService)
	return usersService, mockRepo, mockNotifier
}

func TestUsersService_UpdateUser(t *testing.T) {
	userID := 1
	req := &requests.User{
		Mail:  "new.email@example.com",
		Phone: "+48987654321",
	}
	userFromDB := &ent.User{
		ID:    userID,
		Mail:  "old.email@example.com",
		Phone: "+48123456789",
		Edges: ent.UserEdges{
			UserSettings:     &ent.UserSettings{},
			VerificationCode: &ent.VerificationCode{},
			ActivationCode:   &ent.ActivationCode{},
		},
	}
	updatedUserFromDB := &ent.User{
		ID:    userID,
		Mail:  req.Mail,
		Phone: req.Phone,
		Edges: ent.UserEdges{
			UserSettings:     &ent.UserSettings{},
			VerificationCode: &ent.VerificationCode{},
			ActivationCode:   &ent.ActivationCode{},
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
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("UpdateUser", mock.Anything, userFromDB, req).Return(updatedUserFromDB, nil).Once()
				notifier.On("CreateVerificationEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				notifier.On("CreateRegistrationEvent", mock.Anything, mock.Anything).Return(nil).Once()
				notifier.On("CreateNotificationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user",
		},
		{
			name: "Failure - Update Fails",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("UpdateUser", mock.Anything, userFromDB, req).Return(nil, errors.New("db error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to update user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			usersService, mockRepo, mockNotifier := newTestUsersService(t)
			tc.setupMock(mockRepo, mockNotifier)

			updatedUser, err := usersService.UpdateUser(context.Background(), userID, req)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				assert.Nil(t, updatedUser)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, updatedUser)
				assert.Equal(t, req.Mail, updatedUser.Mail)
			}
			mockRepo.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}

func TestUsersService_UpdateDetails(t *testing.T) {
	userID := 1
	req := &requests.Details{
		FirstName: "John",
		LastName:  "Doe",
	}
	userFromDB := &ent.User{ID: userID, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("UpdateUsersDetails", mock.Anything, userFromDB, req).Return(&ent.User{Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}, nil).Once()
				notifier.On("CreateNotificationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			usersService, mockRepo, mockNotifier := newTestUsersService(t)
			tc.setupMock(mockRepo, mockNotifier)

			_, err := usersService.UpdateDetails(context.Background(), userID, req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUsersService_UpdateSettings(t *testing.T) {
	userID := 1
	req := &requests.Settings{
		NightMode: true,
		TwoFactor: true,
	}
	userFromDB := &ent.User{ID: userID, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("UpdateUsersSettings", mock.Anything, userFromDB, req).Return(&ent.User{Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}, nil).Once()
				notifier.On("CreateNotificationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			usersService, mockRepo, mockNotifier := newTestUsersService(t)
			tc.setupMock(mockRepo, mockNotifier)

			_, err := usersService.UpdateSettings(context.Background(), userID, req)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUsersService_RemoveAccount(t *testing.T) {
	userID := 1
	userFromDB := &ent.User{ID: userID, Edges: ent.UserEdges{UserSettings: &ent.UserSettings{}}}

	testCases := []struct {
		name      string
		setupMock func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier)
		expectErr bool
		errMsg    string
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("RemoveAccount", mock.Anything, userFromDB).Return(nil).Once()
				notifier.On("CreateNotificationEvent", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - User Not Found",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
			},
			expectErr: true,
			errMsg:    "failed to fetch user",
		},
		{
			name: "Failure - Remove Fails in Repo",
			setupMock: func(repo *mocks.MockUsersRepository, notifier *mocks.MockEventNotifier) {
				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("RemoveAccount", mock.Anything, userFromDB).Return(errors.New("db error")).Once()
			},
			expectErr: true,
			errMsg:    "failed to remove user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			usersService, mockRepo, mockNotifier := newTestUsersService(t)
			tc.setupMock(mockRepo, mockNotifier)

			err := usersService.RemoveAccount(context.Background(), userID)

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

// func TestUsersService_GetUserPublic(t *testing.T) {
// 	userID := 1
// 	userFromDB := &ent.User{ID: userID, Mail: "public@example.com"}

// 	testCases := []struct {
// 		name      string
// 		setupMock func(repo *mocks.MockUsersRepository)
// 		expectErr bool
// 	}{
// 		{
// 			name: "Success - User Found",
// 			setupMock: func(repo *mocks.MockUsersRepository) {
// 				repo.On("GetUserByID", mock.Anything, userID).Return(userFromDB, nil).Once()
// 			},
// 			expectErr: false,
// 		},
// 		{
// 			name: "Failure - User Not Found",
// 			setupMock: func(repo *mocks.MockUsersRepository) {
// 				repo.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
// 			},
// 			expectErr: true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			usersService, mockRepo, _ := newTestUsersService(t)
// 			tc.setupMock(mockRepo)

// 			user, err := usersService.GetUserPublic(context.Background(), userID)

// 			if tc.expectErr {
// 				require.Error(t, err)
// 				assert.Nil(t, user)
// 			} else {
// 				require.NoError(t, err)
// 				assert.NotNil(t, user)
// 				assert.Equal(t, userID, user.ID)
// 			}
// 			mockRepo.AssertExpectations(t)
// 		})
// 	}
// }
