package infrastructure

import (
	"fmt"
	"sync"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"users-service/app/models/handlers"
	"users-service/tests/mocks"
)

func newTestConsumerManager(t *testing.T) (*ConsumerManager, *mocks.MockConsumerWrapper) {
	t.Helper()
	logger := zerolog.Nop()
	manager := NewConsumerManager(logger)
	consumer := new(mocks.MockConsumerWrapper)
	return manager, consumer
}

func TestConsumerManager_Register(t *testing.T) {
	newWrappedMock := func() *handlers.ConsumerWrapper {
		return handlers.NewConsumerWrapper(new(mocks.MockConsumerWrapper), func() {})
	}

	testCases := []struct {
		name        string
		topic       string
		consumer    *handlers.ConsumerWrapper
		setupMock   func(m *mocks.MockConsumerWrapper)
		expectErr   bool
		errContains string
	}{
		{
			name:      "Success",
			topic:     "test-topic",
			consumer:  newWrappedMock(),
			setupMock: func(m *mocks.MockConsumerWrapper) {},
			expectErr: false,
		},
		{
			name:      "Overwrite Existing Consumer",
			topic:     "test-topic",
			consumer:  newWrappedMock(),
			setupMock: func(m *mocks.MockConsumerWrapper) {},
			expectErr: false,
		},
		{
			name:        "Validation Error - Empty Topic",
			topic:       "",
			consumer:    newWrappedMock(),
			setupMock:   func(m *mocks.MockConsumerWrapper) {},
			expectErr:   true,
			errContains: "topic cannot be empty",
		},
		{
			name:        "Validation Error - Nil Consumer",
			topic:       "test-topic",
			consumer:    nil,
			setupMock:   func(m *mocks.MockConsumerWrapper) {},
			expectErr:   true,
			errContains: "consumer cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, _ := newTestConsumerManager(t)

			if tc.name == "Overwrite Existing Consumer" {
				err := manager.Register(tc.topic, newWrappedMock())
				require.NoError(t, err)
			}

			err := manager.Register(tc.topic, tc.consumer)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
				_, exists := manager.GetConsumer(tc.topic)
				assert.True(t, exists, "consumer should be registered for topic %q", tc.topic)
			}
		})
	}
}

func TestConsumerManager_PingAll(t *testing.T) {
	var nilString *string
	testCases := []struct {
		name        string
		setupMock   func(m *mocks.MockConsumerWrapper)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				m.On("GetMetadata", nilString, true, 1000).Return(&kafka.Metadata{}, nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Consumer Error - Retriable",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				retriableKafkaErr := kafka.NewError(kafka.ErrTimedOut, "mock timed out", true)
				m.On("GetMetadata", nilString, true, 1000).Return(nil, retriableKafkaErr).Once()
			},
			expectErr:   true,
			errContains: "consumer for topic 'test-topic' is unhealthy: Fatal error: mock timed out",
		},
		{
			name: "Consumer Error - Non-Retriable",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				nonRetriableKafkaErr := kafka.NewError(kafka.ErrBrokerNotAvailable, "broker not available", false)
				m.On("GetMetadata", nilString, true, 1000).Return(nil, nonRetriableKafkaErr).Once()
			},
			expectErr:   true,
			errContains: "consumer for topic 'test-topic' is unhealthy: broker not available",
		},
		{
			name:      "No Consumers",
			setupMock: func(m *mocks.MockConsumerWrapper) {},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, mockConsumer := newTestConsumerManager(t)
			if tc.name != "No Consumers" {
				wrappedConsumer := handlers.NewConsumerWrapper(mockConsumer, func() {})
				err := manager.Register("test-topic", wrappedConsumer)
				require.NoError(t, err)
			}
			tc.setupMock(mockConsumer)

			err := manager.PingAll()

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockConsumer.AssertExpectations(t)
		})
	}
}

func TestConsumerManager_ConcurrentAccess(t *testing.T) {
	manager, _ := newTestConsumerManager(t)
	const numGoroutines = 10
	var nilString *string
	consumers := make([]*mocks.MockConsumerWrapper, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		consumers[i] = new(mocks.MockConsumerWrapper)
		consumers[i].On("GetMetadata", nilString, mock.Anything, mock.Anything).Return(&kafka.Metadata{}, nil).Times(numGoroutines)
		topic := fmt.Sprintf("topic-%d", i)

		wrappedConsumer := handlers.NewConsumerWrapper(consumers[i], func() {})
		err := manager.Register(topic, wrappedConsumer)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	var errorsMu sync.Mutex
	var testErrors []error

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := manager.PingAll(); err != nil {
				errorsMu.Lock()
				testErrors = append(testErrors, fmt.Errorf("PingAll() failed for goroutine %d: %v", i, err))
				errorsMu.Unlock()
				return
			}
		}(i)
	}
	wg.Wait()

	for _, err := range testErrors {
		t.Error(err)
	}

	for i := 0; i < numGoroutines; i++ {
		topic := fmt.Sprintf("topic-%d", i)
		_, exists := manager.GetConsumer(topic)
		assert.True(t, exists, "consumer should be registered for topic %q", topic)
	}

	for i := 0; i < numGoroutines; i++ {
		consumers[i].AssertExpectations(t)
	}
}
