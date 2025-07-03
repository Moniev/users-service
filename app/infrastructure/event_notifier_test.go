//go:build unit

package infrastructure

import (
	"errors"
	"testing"
	"time"
	"users-service/app/models/ent"
	"users-service/tests/mocks"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var _ KafkaProducerInterface = (*mocks.MockKafkaProducer)(nil)

func newTestNotifier(t *testing.T) (*EventNotifier, *mocks.MockKafkaProducer) {
	mockProducer := new(mocks.MockKafkaProducer)
	notifier := NewEventNotifier(mockProducer, zerolog.Nop(), "test-topic", time.Second*15)
	require.NotNil(t, notifier)
	return notifier, mockProducer
}

func TestCreateRegistrationEvent(t *testing.T) {
	baseUser := &ent.User{
		ID:   1,
		Mail: "test@example.com",
		Edges: ent.UserEdges{
			UserSettings: &ent.UserSettings{
				UUID:  "user-uuid-123",
				Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
			},
		},
	}
	baseCode := &ent.ActivationCode{Code: "123456"}

	testCases := []struct {
		name        string
		user        *ent.User
		code        *ent.ActivationCode
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success",
			user: baseUser,
			code: baseCode,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Producer Error",
			user: baseUser,
			code: baseCode,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(errors.New("kafka is down")).Once()
			},
			expectErr:   true,
			errContains: "kafka is down",
		},
		{
			name:        "Validation Error - Missing UserSettings",
			user:        &ent.User{ID: 1},
			code:        baseCode,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings not loaded for user",
		},
		{
			name: "Validation Error - Missing Owner in UserSettings",
			user: &ent.User{
				ID: 1,
				Edges: ent.UserEdges{
					UserSettings: &ent.UserSettings{},
				},
			},
			code:        baseCode,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings owner not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateRegistrationEvent(tc.user, tc.code)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateLoginEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID:  "user-uuid-123",
		Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
	}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		loginMethod string
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:        "Success",
			settings:    baseSettings,
			loginMethod: "password",
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name:        "Validation Error - Missing Owner",
			settings:    &ent.UserSettings{},
			loginMethod: "password",
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings owner not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateLoginEvent(tc.settings, tc.loginMethod)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateSecondFactorEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID: "user-uuid-123",
		Edges: ent.UserSettingsEdges{
			Owner:              &ent.User{ID: 1},
			SecondFactorTarget: &ent.UserDevice{Token: "device-token"},
		},
	}
	baseCode := &ent.SecondFactorCode{Code: "123456"}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		code        *ent.SecondFactorCode
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:     "Success",
			settings: baseSettings,
			code:     baseCode,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Validation Error - Missing SecondFactorTarget",
			settings: &ent.UserSettings{
				Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
			},
			code:        baseCode,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "2fa target device not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateSecondFactorEvent(tc.settings, tc.code)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateVerificationEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID:  "user-uuid-123",
		Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
	}
	baseCode := &ent.VerificationCode{Code: "123456"}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		phone       string
		code        *ent.VerificationCode
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:     "Success",
			settings: baseSettings,
			phone:    "+123456789",
			code:     baseCode,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name:        "Validation Error - Missing Owner",
			settings:    &ent.UserSettings{},
			phone:       "+123456789",
			code:        baseCode,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings owner not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateVerificationEvent(tc.settings, tc.phone, tc.code)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateNotificationEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID:  "user-uuid-123",
		Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
	}
	devices := []*ent.UserDevice{{Token: "token1"}, {Token: "token2"}}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		devices     []*ent.UserDevice
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:     "Success",
			settings: baseSettings,
			devices:  devices,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name:        "Validation Error - Missing Owner",
			settings:    &ent.UserSettings{},
			devices:     devices,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings owner not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateNotificationEvent(tc.settings, tc.devices)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateUserActionEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID:  "user-uuid-123",
		Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
	}
	action := &ent.UserAction{Action: "test.action", Details: "some details"}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		action      *ent.UserAction
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:     "Success",
			settings: baseSettings,
			action:   action,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name:        "Validation Error - Missing Owner",
			settings:    &ent.UserSettings{},
			action:      action,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "user settings owner not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateUserActionEvent(tc.settings, tc.action)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestCreateResetPasswordEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID: "user-uuid-123",
		Edges: ent.UserSettingsEdges{
			Owner:              &ent.User{ID: 1, Phone: "+123456789"},
			SecondFactorTarget: &ent.UserDevice{Token: "device-token"},
		},
	}
	baseCode := &ent.ResetCode{Code: "654321"}

	testCases := []struct {
		name        string
		settings    *ent.UserSettings
		code        *ent.ResetCode
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name:     "Success",
			settings: baseSettings,
			code:     baseCode,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Validation Error - Missing SecondFactorTarget",
			settings: &ent.UserSettings{
				Edges: ent.UserSettingsEdges{Owner: &ent.User{ID: 1}},
			},
			code:        baseCode,
			setupMock:   func(m *mocks.MockKafkaProducer) {},
			expectErr:   true,
			errContains: "target device for reset not loaded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateResetPasswordEvent(tc.settings, tc.code)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestLogoutEvent(t *testing.T) {
	baseSettings := &ent.UserSettings{
		UUID: "user-uuid-123",
		Edges: ent.UserSettingsEdges{
			Owner:              &ent.User{ID: 1, Phone: "+123456789"},
			SecondFactorTarget: &ent.UserDevice{Token: "device-token"},
		},
	}
	logoutMethod := "logout-method"

	testCases := []struct {
		name         string
		settings     *ent.UserSettings
		logoutMethod string
		setupMock    func(m *mocks.MockKafkaProducer)
		expectErr    bool
		errContains  string
	}{
		{
			name:         "Success",
			settings:     baseSettings,
			logoutMethod: logoutMethod,
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("Produce", mock.AnythingOfType("*kafka.Message"), mock.Anything).Return(nil).Once()
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.CreateLogoutEvent(tc.settings, tc.logoutMethod)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}

func TestPing(t *testing.T) {
	testCases := []struct {
		name        string
		setupMock   func(m *mocks.MockKafkaProducer)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.MockKafkaProducer) {
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, nil).Once()
			},
			expectErr: false,
		},
		{
			name: "Producer Ping Error - Retriable",
			setupMock: func(m *mocks.MockKafkaProducer) {
				retriableKafkaErr := kafka.NewError(kafka.ErrTimedOut, "mock timed out", true)
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, retriableKafkaErr).Once()
			},
			expectErr: false,
		},
		{
			name: "Producer Ping Error - Non-Retriable",
			setupMock: func(m *mocks.MockKafkaProducer) {
				nonRetriableKafkaErr := kafka.NewError(kafka.ErrBrokerNotAvailable, "broker not available", false)
				m.On("GetMetadata", mock.Anything, true, mock.AnythingOfType("int")).Return(nil, nonRetriableKafkaErr).Once()
			},
			expectErr:   true,
			errContains: "kafka producer not connected or unhealthy: broker not available",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notifier, mockProducer := newTestNotifier(t)
			tc.setupMock(mockProducer)

			err := notifier.Ping()

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockProducer.AssertExpectations(t)
		})
	}
}
