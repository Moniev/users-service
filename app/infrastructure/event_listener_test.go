//go:build unit

package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"
	"users-service/app/models/events"
	"users-service/app/models/utils"
	"users-service/tests/mocks"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var listenerTopic string = "test-topic"

var _ utils.ConsumerWrapperInterface = (*mocks.MockConsumerWrapper)(nil)

func newTestEventListener(t *testing.T) (*EventListener, *mocks.MockConsumerWrapper) {
	mockConsumer := &mocks.MockConsumerWrapper{}
	logger := zerolog.New(zerolog.NewTestWriter(t))
	topic := "test-topic"
	maxInactivity := 30 * time.Second
	kafkaPingTimeout := 5 * time.Second
	listener := NewEventListener(mockConsumer, logger, topic, maxInactivity, kafkaPingTimeout)
	return listener, mockConsumer
}

func TestEventListener_HandleUserActionEvent(t *testing.T) {
	listener, _ := newTestEventListener(t)
	ctx := context.Background()

	testCases := []struct {
		name        string
		event       *events.UserActionEvent
		expectErr   bool
		errContains string
	}{
		{
			name: "Success - Valid Event",
			event: &events.UserActionEvent{
				BaseEvent: events.BaseEvent{
					EventID:   "123",
					EventType: "user.action.test",
					UserID:    1,
				},
				Action: "user_clicked_button",
			},
			expectErr: false,
		},
		{
			name:        "Failure - Nil Event",
			event:       nil,
			expectErr:   true,
			errContains: "invalid user action event",
		},
		{
			name: "Failure - Empty Action",
			event: &events.UserActionEvent{
				BaseEvent: events.BaseEvent{
					EventID:   "123",
					EventType: "user.action.test",
					UserID:    1,
				},
				Action: "",
			},
			expectErr:   true,
			errContains: "invalid user action event",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := listener.HandleUserActionEvent(ctx, tc.event)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEventListener_Ping(t *testing.T) {
	testCases := []struct {
		name        string
		setupMock   func(m *mocks.MockConsumerWrapper)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success - Active and Connected",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Inactive Listener Loop",
			setupMock: func(m *mocks.MockConsumerWrapper) {
			},
			expectErr:   true,
			errContains: "consumer listener loop inactive for too long",
		},
		{
			name: "Failure - Kafka Disconnected (Non-Retriable)",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				nonRetriableErr := kafka.NewError(kafka.ErrBrokerNotAvailable, "broker offline", false)
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, nonRetriableErr).Once()
			},
			expectErr:   true,
			errContains: "kafka consumer not connected or unhealthy",
		},
		{
			name: "Success - Kafka Temporary Issue (Retriable)",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				retriableErr := kafka.NewError(kafka.ErrTimedOut, "request timed out", true)
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, retriableErr).Once()
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listener, mockConsumer := newTestEventListener(t)
			tc.setupMock(mockConsumer)

			if tc.name == "Failure - Inactive Listener Loop" {
				listener.mu.Lock()
				listener.lastActivityTime = time.Now().Add(-listener.MaxInactivity - time.Second)
				listener.mu.Unlock()
			}

			err := listener.Ping()

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

func TestEventListener_Listen(t *testing.T) {
	testCases := []struct {
		name              string
		setupMock         func(m *mocks.MockConsumerWrapper)
		stopListenAfter   time.Duration
		expectErr         bool
		errContains       string
		expectedCommitMsg bool
	}{
		{
			name: "Graceful Shutdown - Context Done",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 50 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "Kafka ReadMessage Timeout - Loop Continues",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 150 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "Kafka ReadMessage Error - Loop Continues",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				someKafkaErr := kafka.NewError(kafka.ErrUnknown, "some kafka error", false)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, someKafkaErr).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 150 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "Unmarshal BaseEvent Error - Skips Message",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				badJSONMsg := &kafka.Message{Value: []byte("{invalid json}"), TopicPartition: kafka.TopicPartition{Topic: &listenerTopic, Partition: 0}}
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(badJSONMsg, nil).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 150 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "Unknown EventType - Skips Message",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				unknownEventMsg := &kafka.Message{Value: []byte(`{"event_type": "unknown.event"}`), TopicPartition: kafka.TopicPartition{Topic: &listenerTopic, Partition: 0}}
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(unknownEventMsg, nil).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 150 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "HandleUserActionEvent Error - Skips Message",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				msgWithBadAction := &kafka.Message{Value: []byte(`{"event_type": "user.action.test", "action": ""}`), TopicPartition: kafka.TopicPartition{Topic: &listenerTopic, Partition: 0}}
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(msgWithBadAction, nil).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter: 150 * time.Millisecond,
			expectErr:       false,
		},
		{
			name: "CommitMessage Error - Loop Continues",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				validMsg := &kafka.Message{Value: []byte(`{"event_type": "user.action.login", "event_id": "1", "user_id": 10, "action": "login_successful"}`), TopicPartition: kafka.TopicPartition{Topic: &listenerTopic, Partition: 0}}
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(validMsg, nil).Once()
				m.On("CommitMessage", mock.AnythingOfType("*kafka.Message")).Return(nil, errors.New("commit failed")).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter:   150 * time.Millisecond,
			expectErr:         false,
			expectedCommitMsg: true,
		},
		{
			name: "Success - Process and Commit",
			setupMock: func(m *mocks.MockConsumerWrapper) {
				validMsg := &kafka.Message{Value: []byte(`{"event_type": "user.action.login", "event_id": "1", "user_id": 10, "action": "login_successful"}`), TopicPartition: kafka.TopicPartition{Topic: &listenerTopic, Partition: 0}}
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(validMsg, nil).Once()
				m.On("CommitMessage", mock.AnythingOfType("*kafka.Message")).Return([]kafka.TopicPartition{}, nil).Once()
				kafkaErr := kafka.NewError(kafka.ErrTimedOut, "read timeout for shutdown", true)
				m.On("ReadMessage", mock.AnythingOfType("time.Duration")).Return(nil, kafkaErr).Maybe()
			},
			stopListenAfter:   150 * time.Millisecond,
			expectErr:         false,
			expectedCommitMsg: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listener, mockConsumerWrapper := newTestEventListener(t)
			tc.setupMock(mockConsumerWrapper)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				time.Sleep(tc.stopListenAfter)
				cancel()
			}()

			err := listener.Listen(ctx)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			if tc.expectedCommitMsg {
				mockConsumerWrapper.AssertCalled(t, "CommitMessage", mock.AnythingOfType("*kafka.Message"))
			} else if tc.name != "Graceful Shutdown - Context Done" {
				mockConsumerWrapper.AssertNotCalled(t, "CommitMessage", mock.Anything)
			}

			mockConsumerWrapper.AssertExpectations(t)
		})
	}
}
