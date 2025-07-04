package infrastructure

import (
	"encoding/json"
	"errors"
	"time"
	"users-service/app/models/ent"
	"users-service/app/models/events"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type EventNotifier struct {
	KafkaProducer    KafkaProducerInterface
	Topic            string
	Logger           zerolog.Logger
	KafkaPingTimeout time.Duration
}

type KafkaProducerInterface interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
}

type EventNotifierInterface interface {
	CreateRegistrationEvent(user *ent.User, activationCode *ent.ActivationCode) error
	CreateLoginEvent(settings *ent.UserSettings, loginMethod string) error
	CreateLogoutEvent(settings *ent.UserSettings, logoutMethod string) error

	CreateSecondFactorEvent(settings *ent.UserSettings, secondFactor *ent.SecondFactorCode) error
	CreateVerificationEvent(settings *ent.UserSettings, phone string, verificationCode *ent.VerificationCode) error
	CreateResetPasswordEvent(settings *ent.UserSettings, resetCode *ent.ResetCode) error

	CreateNotificationEvent(settings *ent.UserSettings, devices []*ent.UserDevice) error
	CreateUserActionEvent(settings *ent.UserSettings, action *ent.UserAction) error

	Produce(key []byte, value []byte) error
	Ping() error
}

var _ EventNotifierInterface = (*EventNotifier)(nil)

func NewEventNotifier(
	kafkaProducer KafkaProducerInterface,
	logger zerolog.Logger,
	topic string,
	timeout time.Duration) *EventNotifier {
	logger.Info().Str("topic", topic).Dur("timeout", timeout).Msg("Initializing new EventNotifier")
	return &EventNotifier{
		KafkaProducer:    kafkaProducer,
		Topic:            topic,
		Logger:           logger,
		KafkaPingTimeout: timeout,
	}
}

func (n *EventNotifier) Produce(key []byte, value []byte) error {
	n.Logger.Debug().Str("topic", n.Topic).Int("keyLength", len(key)).Int("valueLength", len(value)).Msg("Attempting to produce Kafka message")
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &n.Topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
		Timestamp:      time.Now().UTC(),
	}

	err := n.KafkaProducer.Produce(message, nil)
	if err != nil {
		n.Logger.Error().Err(err).Str("topic", n.Topic).Msg("Failed to produce Kafka message")
		return err
	}

	n.Logger.Debug().Str("topic", n.Topic).Msg("Kafka message produced successfully")
	return nil
}

func (n *EventNotifier) createAndProduceEvent(event interface{}, settings *ent.UserSettings, eventType string) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}

	n.Logger.Debug().Str("event_type", eventType).Int("user_id", userID).Msg("Creating and producing event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Str("event_type", eventType).Msg("User settings owner not loaded, cannot produce event")
		return errors.New("user settings owner not loaded")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		n.Logger.Error().Err(err).Str("event_type", eventType).Int("user_id", userID).Msg("Failed to marshal event payload to JSON")
		return err
	}
	n.Logger.Debug().Str("event_type", eventType).Int("user_id", userID).Int("payloadLength", len(payload)).Msg("Event payload marshaled")

	if err := n.Produce([]byte(settings.UUID), payload); err != nil {
		n.Logger.Error().Err(err).Str("event_type", eventType).Int("user_id", userID).Msg("Failed to produce event to Kafka")
		return err
	}

	n.Logger.Info().Str("event_type", eventType).Int("user_id", userID).Msg("Event created and produced successfully")
	return nil
}

func (n *EventNotifier) CreateRegistrationEvent(user *ent.User, activationCode *ent.ActivationCode) error {
	n.Logger.Info().Int("user_id", user.ID).Msg("Attempting to create registration event")

	if user.Edges.UserSettings == nil {
		n.Logger.Error().Int("user_id", user.ID).Msg("Cannot create registration event: User.Edges.UserSettings is not loaded.")
		return errors.New("user settings not loaded for user")
	}
	n.Logger.Debug().Int("user_id", user.ID).Msg("User settings loaded for registration event")

	event := events.RegistrationEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.registered",
			Timestamp: time.Now().UTC(),
			UserID:    user.ID,
		},
		Email:          user.Mail,
		ActivationCode: activationCode.Code,
		ExpiresAt:      activationCode.ExpiresAt,
	}
	n.Logger.Debug().Int("user_id", user.ID).Str("eventID", event.EventID).Msg("Registration event struct created")

	return n.createAndProduceEvent(event, user.Edges.UserSettings, event.EventType)
}

func (n *EventNotifier) CreateSecondFactorEvent(settings *ent.UserSettings, secondFactor *ent.SecondFactorCode) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Msg("Attempting to create second factor event")

	if settings.Edges.SecondFactorTarget == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create 2FA event: UserSettings.Edges.SecondFactorTarget is not loaded.")
		return errors.New("2fa target device not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("2FA target device loaded for second factor event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create 2FA event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for second factor event")

	event := events.SecondFactorEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.secondfactor.setup",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		SecondFactorCode: secondFactor.Code,
		ExpiresAt:        secondFactor.ExpiresAt,
		TargetDevice:     settings.Edges.SecondFactorTarget.Token,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Second factor event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateLoginEvent(settings *ent.UserSettings, loginMethod string) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Str("loginMethod", loginMethod).Msg("Attempting to create login event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create login event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for login event")

	event := events.LoginEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.login.success",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		LoginMethod: loginMethod,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Login event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateVerificationEvent(settings *ent.UserSettings, phone string, verificationCode *ent.VerificationCode) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Str("phone", phone).Msg("Attempting to create verification event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create verification event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for verification event")

	event := events.VerificationEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.phone.verification",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		PhoneNumber:      phone,
		VerificationCode: verificationCode.Code,
		ExpiresAt:        verificationCode.ExpiresAt,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Verification event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateNotificationEvent(settings *ent.UserSettings, devices []*ent.UserDevice) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Int("deviceCount", len(devices)).Msg("Attempting to create notification event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create notification event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for notification event")

	deviceTokens := make([]string, len(devices))
	for i, d := range devices {
		deviceTokens[i] = d.Token
	}
	n.Logger.Debug().Int("user_id", userID).Strs("deviceTokens", deviceTokens).Msg("Extracted target device tokens")

	event := events.NotificationEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.notification",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		TargetDevices: deviceTokens,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Notification event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateUserActionEvent(settings *ent.UserSettings, action *ent.UserAction) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Str("actionType", action.Action).Msg("Attempting to create user action event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create user action event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for user action event")

	event := events.UserActionEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.action." + action.Action,
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		Action:  action.Action,
		Details: action.Details,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("User action event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateResetPasswordEvent(settings *ent.UserSettings, resetCode *ent.ResetCode) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Msg("Attempting to create reset password event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create reset password event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for reset password event")

	if settings.Edges.SecondFactorTarget == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create reset password event: UserSettings.Edges.SecondFactorTarget is not loaded.")
		return errors.New("target device for reset not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("Second factor target loaded for reset password event")

	if settings.Edges.Owner.Phone == "" {
		n.Logger.Warn().Int("user_id", userID).Msg("User has no phone number for password reset event. Event will proceed without phone number.")
	}
	n.Logger.Debug().Int("user_id", userID).Str("userPhone", settings.Edges.Owner.Phone).Msg("Checking user phone for reset event")

	event := events.ResetPasswordEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.reset.password",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		PhoneNumber:  settings.Edges.Owner.Phone,
		ResetCode:    resetCode.Code,
		ExpiresAt:    resetCode.ExpiresAt,
		TargetDevice: settings.Edges.SecondFactorTarget.Token,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Reset password event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateLogoutEvent(settings *ent.UserSettings, logoutMethod string) error {
	userID := 0
	if settings.Edges.Owner != nil {
		userID = settings.Edges.Owner.ID
	}
	n.Logger.Info().Int("user_id", userID).Str("logoutMethod", logoutMethod).Msg("Attempting to create logout event")

	if settings.Edges.Owner == nil {
		n.Logger.Error().Int("user_id", userID).Msg("Cannot create logout event: User settings owner not loaded.")
		return errors.New("user settings owner not loaded")
	}
	n.Logger.Debug().Int("user_id", userID).Msg("User settings owner loaded for logout event")

	event := events.LogoutEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.logout.success",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		LogoutMethod: logoutMethod,
	}
	n.Logger.Debug().Int("user_id", userID).Str("eventID", event.EventID).Msg("Logout event struct created")

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) Ping() error {
	n.Logger.Info().Msg("Performing Kafka producer health check")
	timeoutMs := int(n.KafkaPingTimeout / time.Millisecond)

	_, err := n.KafkaProducer.GetMetadata(nil, true, timeoutMs)

	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok {
			if kafkaErr.IsTimeout() {
				n.Logger.Warn().Err(err).Msg("Kafka producer connection check returned a timeout error. Still considered healthy for now.")
				return nil
			}

			if kafkaErr.IsRetriable() {
				n.Logger.Warn().Err(err).Msg("Kafka producer connection check returned a retriable error. Still considered healthy for now.")
				return nil
			}
		}

		n.Logger.Error().Err(err).Msg("Kafka producer failed to get metadata during health check (connection issue)")
		return errors.New("kafka producer not connected or unhealthy: " + err.Error())
	}

	n.Logger.Info().Msg("Kafka producer health check passed successfully")
	return nil
}
