//go:build unit

package services

import (
	"context"
	"errors"
	"testing"
	"users-service/tests/mocks"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestDiagnosticsService(t *testing.T) (*DiagnosticsService, *mocks.MockUsersRepository, *mocks.MockEventNotifier) {
	mockRepo := new(mocks.MockUsersRepository)
	eventNotifier := new(mocks.MockEventNotifier)

	cacheStore := new(mocks.MockCacheStore)

	eventListener := new(mocks.MockEventListener)

	logger := zerolog.Nop()

	authService := NewDiagnosticsService(
		mockRepo,
		cacheStore,
		eventListener,
		eventNotifier,
		logger,
	)
	require.NotNil(t, authService)
	return authService, mockRepo, eventNotifier
}

func TestDiagnosticsService_CheckHealth(t *testing.T) {
	testCases := []struct {
		name        string
		setupMock   func(repo *mocks.MockUsersRepository)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success - Healthy Database",
			setupMock: func(repo *mocks.MockUsersRepository) {
				repo.On("Ping").Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Unhealthy Database",
			setupMock: func(repo *mocks.MockUsersRepository) {
				repo.On("Ping").Return(errors.New("database connection failed")).Once()
			},
			expectErr:   true,
			errContains: "database unhealthy: database connection failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, mockRepo, _ := newTestDiagnosticsService(t)
			tc.setupMock(mockRepo)

			err := service.CheckHealth(context.Background())

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiagnosticsService_CheckReadiness(t *testing.T) {
	testCases := []struct {
		name        string
		setupMocks  func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success - All Components Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				listener.On("Ping").Return(nil).Once()
				notifier.On("Ping").Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - UsersRepository Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(errors.New("database connection failed")).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Maybe()
				listener.On("Ping").Return(nil).Maybe()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "users repository not ready: database connection failed",
		},
		{
			name: "Failure - CacheStore Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("redis connection failed"))
				cache.On("Ping", mock.Anything).Return(cmd).Once()
				listener.On("Ping").Return(nil).Maybe()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "cache store not ready: redis connection failed",
		},
		{
			name: "Failure - EventListener Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				listener.On("Ping").Return(errors.New("kafka consumer not connected")).Once()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "event listener not ready: kafka consumer not connected",
		},
		{
			name: "Failure - EventNotifier Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, listener *mocks.MockEventListener, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				listener.On("Ping").Return(nil).Once()
				notifier.On("Ping").Return(errors.New("kafka producer not connected")).Once()
			},
			expectErr:   true,
			errContains: "event notifier not ready: kafka producer not connected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, mockRepo, mockNotifier := newTestDiagnosticsService(t)
			mockCache := service.CacheStore.(*mocks.MockCacheStore)
			mockListener := service.EventListener.(*mocks.MockEventListener)
			tc.setupMocks(mockRepo, mockCache, mockListener, mockNotifier)

			err := service.CheckReadiness(context.Background())

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockListener.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
