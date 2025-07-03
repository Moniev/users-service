package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"users-service/app/models/events"
	"users-service/app/models/utils"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type EventListener struct {
	Consumer *utils.ConsumerWrapper
	Logger   zerolog.Logger
	Topic    string
}

type EventListenerInterface interface {
	HandleUserActionEvent(ctx context.Context, event *events.UserActionEvent) error
}

var _ EventListenerInterface = (*EventListener)(nil)

func NewEventListener(
	consumer *utils.ConsumerWrapper,
	logger zerolog.Logger,
	topic string) *EventListener {

	return &EventListener{
		Consumer: consumer,
		Logger:   logger,
		Topic:    topic,
	}
}

func (l *EventListener) HandleUserActionEvent(ctx context.Context, event *events.UserActionEvent) error {
	if event == nil || event.Action == "" {
		l.Logger.Error().Msg("Invalid user action event: missing required fields")
		return errors.New("invalid user action event")
	}

	l.Logger.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Int("user_id", event.UserID).
		Str("action", event.Action).
		Msg("Processed user action event")
	return nil
}

func (l *EventListener) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			l.Logger.Info().Str("topic", l.Topic).Msg("Stopping event listener")
			return nil
		default:
			msg, err := l.Consumer.Consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				l.Logger.Error().Err(err).Msg("Failed to read Kafka message")
				continue
			}

			var baseEvent events.BaseEvent
			if err := json.Unmarshal(msg.Value, &baseEvent); err != nil {
				l.Logger.Error().Err(err).Str("topic", *msg.TopicPartition.Topic).Msg("Failed to unmarshal base event")
				continue
			}

			var handlerErr error
			switch baseEvent.EventType {
			case "user.action":
				if baseEvent.EventType != "" && baseEvent.EventType[:12] == "user.action." {
					var event events.UserActionEvent
					if err := json.Unmarshal(msg.Value, &event); err != nil {
						l.Logger.Error().Err(err).Str("event_type", baseEvent.EventType).Msg("Failed to unmarshal event")
						continue
					}

					handlerErr = l.HandleUserActionEvent(ctx, &event)
				} else {
					l.Logger.Warn().Str("event_type", baseEvent.EventType).Msg("Unknown event type received")
					continue
				}
			}

			if handlerErr != nil {
				l.Logger.Error().Err(handlerErr).Str("event_type", baseEvent.EventType).Msg("Failed to handle event")
				continue
			}

			_, err = l.Consumer.Consumer.CommitMessage(msg)
			if err != nil {
				l.Logger.Error().Err(err).Str("topic", *msg.TopicPartition.Topic).Msg("Failed to commit offset")
			} else {
				l.Logger.Debug().
					Str("topic", *msg.TopicPartition.Topic).
					Int32("partition", msg.TopicPartition.Partition).
					Int64("offset", int64(msg.TopicPartition.Offset)).
					Msg("Offset committed")
			}
		}
	}
}
