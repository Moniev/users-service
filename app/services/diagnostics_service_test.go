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

func newTestDiagnosticsService(t *testing.T) (
	*DiagnosticsService,
	*mocks.MockUsersRepository,
	*mocks.MockEventNotifier,
	*mocks.MockConsumerManager,
) {
	mockRepo := new(mocks.MockUsersRepository)
	eventNotifier := new(mocks.MockEventNotifier)
	cacheStore := new(mocks.MockCacheStore)
	consumerManager := new(mocks.MockConsumerManager)

	logger := zerolog.Nop()

	service := NewDiagnosticsService(
		mockRepo,
		cacheStore,
		eventNotifier,
		consumerManager,
		logger,
	)
	require.NotNil(t, service)
	return service, mockRepo, eventNotifier, consumerManager
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
			service, mockRepo, _, _ := newTestDiagnosticsService(t)
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
		setupMocks  func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success - All Components Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				manager.On("PingAll").Return(nil).Once()
				notifier.On("Ping").Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - UsersRepository Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(errors.New("database connection failed")).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Maybe()
				manager.On("PingAll").Return(nil).Maybe()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "users repository not ready: database connection failed",
		},
		{
			name: "Failure - CacheStore Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(errors.New("redis connection failed"))
				cache.On("Ping", mock.Anything).Return(cmd).Once()
				manager.On("PingAll").Return(nil).Maybe()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "cache store not ready: redis connection failed",
		},
		{
			name: "Failure - ConsumerManager Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				manager.On("PingAll").Return(errors.New("kafka consumers not connected")).Once()
				notifier.On("Ping").Return(nil).Maybe()
			},
			expectErr:   true,
			errContains: "consumers not ready: kafka consumers not connected",
		},
		{
			name: "Failure - EventNotifier Not Ready",
			setupMocks: func(repo *mocks.MockUsersRepository, cache *mocks.MockCacheStore, manager *mocks.MockConsumerManager, notifier *mocks.MockEventNotifier) {
				repo.On("Ping").Return(nil).Once()
				cache.On("Ping", mock.Anything).Return(&redis.StatusCmd{}, nil).Once()
				manager.On("PingAll").Return(nil).Once()
				notifier.On("Ping").Return(errors.New("kafka producer not connected")).Once()
			},
			expectErr:   true,
			errContains: "event notifier not ready: kafka producer not connected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, mockRepo, mockNotifier, mockManager := newTestDiagnosticsService(t)
			mockCache := service.CacheStore.(*mocks.MockCacheStore)

			tc.setupMocks(mockRepo, mockCache, mockManager, mockNotifier)

			err := service.CheckReadiness(context.Background())

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockManager.AssertExpectations(t)
			mockNotifier.AssertExpectations(t)
		})
	}
}
