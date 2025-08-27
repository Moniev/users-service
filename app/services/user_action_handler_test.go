//go:build unit

package services

import (
	"context"
	"encoding/json"
	"testing"
	"users-service/app/models/ent"
	"users-service/app/models/events"
	"users-service/tests/mocks"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestUserEventHandler(t *testing.T) (*UserActionHandler, *mocks.MockUsersRepository) {
	mockRepo := new(mocks.MockUsersRepository)
	logger := zerolog.Nop()

	handler := NewUserActionHandler(mockRepo, logger)
	require.NotNil(t, handler)

	return handler, mockRepo
}

func TestUserActionHandler_Handle(t *testing.T) {
	userID := 1
	ctx := context.Background()

	validEvent := events.UserActionEvent{
		BaseEvent: events.BaseEvent{
			EventID:   "evt-123",
			EventType: "user.action.login",
			UserID:    1,
		},
		Action:       "action",
		OriginDevice: "original-TOKEN-123",
		Details:      "Something has happend in GAZA",
	}
	validEventPayload, _ := json.Marshal(validEvent)
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

	testCases := []struct {
		name        string
		message     *kafka.Message
		setupMocks  func(repo *mocks.MockUsersRepository)
		expectErr   bool
		errContains string
		expectedLog string
	}{
		{
			name: "Success - Valid User Action Event",
			message: &kafka.Message{
				Value: validEventPayload,
				TopicPartition: kafka.TopicPartition{
					Topic:     &[]string{"user.actions"}[0],
					Partition: 0,
				},
			},
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.On("GetUserPublicByID", mock.Anything, userID).Return(userFromDB, nil).Once()
				repo.On("CreateUserAction", mock.Anything, userFromDB, validEvent.Action, validEvent.OriginDevice, validEvent.Details).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Invalid JSON payload",
			message: &kafka.Message{
				Value: []byte(`{"invalid json`),
			},
			setupMocks:  func(repo *mocks.MockUsersRepository) {},
			expectErr:   true,
			errContains: "failed to unmarshal base event",
		},
		{
			name: "Failure - Mismatched event payload",
			message: &kafka.Message{
				Value: []byte(`{"event_type": "user.action.test", "action": "some_action", "user_id": "not-an-int"}`),
			},
			setupMocks:  func(repo *mocks.MockUsersRepository) {},
			expectErr:   true,
			errContains: "failed to unmarshal base event: json: cannot unmarshal string into Go struct field BaseEvent.user_id of type int",
		},
		{
			name: "Failure - Event missing action field",
			message: &kafka.Message{
				Value: []byte(`{"event_type": "user.action.test", "user_id": 1}`),
			},
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.On("GetUserPublicByID", mock.Anything, userID).Return(userFromDB, nil).Once()
			},
			expectErr:   true,
			errContains: "invalid user action event: missing action",
		},
		{
			name: "Success - Unknown event type is ignored",
			message: &kafka.Message{
				Value: []byte(`{"event_type": "some.other.event"}`),
			},
			setupMocks: func(repo *mocks.MockUsersRepository) {},
			expectErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, mockRepo := newTestUserEventHandler(t)
			tc.setupMocks(mockRepo)

			err := handler.Handle(ctx, tc.message)

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
