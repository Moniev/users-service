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
	return &EventNotifier{
		KafkaProducer:    kafkaProducer,
		Topic:            topic,
		Logger:           logger,
		KafkaPingTimeout: timeout,
	}
}

func (n *EventNotifier) Produce(key []byte, value []byte) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &n.Topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
		Timestamp:      time.Now().UTC(),
	}

	err := n.KafkaProducer.Produce(message, nil)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Failed to produce Kafka message")
		return err
	}

	return nil
}

func (n *EventNotifier) createAndProduceEvent(event interface{}, settings *ent.UserSettings, eventType string) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		n.Logger.Error().Err(err).Str("event_type", eventType).Int("user_id", settings.Edges.Owner.ID).Msg("Failed to marshal event")
		return err
	}

	if err := n.Produce([]byte(settings.UUID), payload); err != nil {
		return err
	}

	n.Logger.Info().Str("event_type", eventType).Int("user_id", settings.Edges.Owner.ID).Msg("Event created and produced")
	return nil
}

func (n *EventNotifier) CreateRegistrationEvent(user *ent.User, activationCode *ent.ActivationCode) error {
	if user.Edges.UserSettings == nil {
		n.Logger.Error().Int("user_id", user.ID).Msg("Cannot create registration event: User.Edges.UserSettings is not loaded.")
		return errors.New("user settings not loaded for user")
	}

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

	return n.createAndProduceEvent(event, user.Edges.UserSettings, event.EventType)
}

func (n *EventNotifier) CreateSecondFactorEvent(settings *ent.UserSettings, secondFactor *ent.SecondFactorCode) error {
	if settings.Edges.SecondFactorTarget == nil {
		n.Logger.Error().Int("user_id", settings.Edges.Owner.ID).Msg("Cannot create 2FA event: UserSettings.Edges.SecondFactorTarget is not loaded.")
		return errors.New("2fa target device not loaded")
	}

	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

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

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateLoginEvent(settings *ent.UserSettings, loginMethod string) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

	event := events.LoginEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.login.success",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		LoginMethod: loginMethod,
	}

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateVerificationEvent(settings *ent.UserSettings, phone string, verificationCode *ent.VerificationCode) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

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

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateNotificationEvent(settings *ent.UserSettings, devices []*ent.UserDevice) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

	deviceTokens := make([]string, len(devices))
	for i, d := range devices {
		deviceTokens[i] = d.Token
	}

	event := events.NotificationEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.notification",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		TargetDevices: deviceTokens,
	}

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateUserActionEvent(settings *ent.UserSettings, action *ent.UserAction) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

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

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateResetPasswordEvent(settings *ent.UserSettings, resetCode *ent.ResetCode) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

	if settings.Edges.SecondFactorTarget == nil {
		n.Logger.Error().Int("user_id", settings.Edges.Owner.ID).Msg("Cannot create reset password event: UserSettings.Edges.SecondFactorTarget is not loaded.")
		return errors.New("target device for reset not loaded")
	}

	if settings.Edges.Owner.Phone == "" {
		n.Logger.Warn().Int("user_id", settings.Edges.Owner.ID).Msg("User has no phone number for password reset event.")
	}

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

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) CreateLogoutEvent(settings *ent.UserSettings, logoutMethod string) error {
	if settings.Edges.Owner == nil {
		return errors.New("user settings owner not loaded")
	}

	event := events.LogoutEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: "user.logout.success",
			Timestamp: time.Now().UTC(),
			UserID:    settings.Edges.Owner.ID,
		},
		LogoutMethod: logoutMethod,
	}

	return n.createAndProduceEvent(event, settings, event.EventType)
}

func (n *EventNotifier) Ping() error {
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

	return nil
}
