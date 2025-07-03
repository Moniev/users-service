package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
	"users-service/app/models/events"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type EventListener struct {
	Consumer         KafkaConsumerInterface
	Logger           zerolog.Logger
	Topic            string
	lastActivityTime time.Time
	mu               sync.RWMutex
	MaxInactivity    time.Duration
	KafkaPingTimeout time.Duration
}

type KafkaConsumerInterface interface {
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error)
}

type EventListenerInterface interface {
	HandleUserActionEvent(ctx context.Context, event *events.UserActionEvent) error
	Listen(ctx context.Context) error
	Ping() error
}

var _ EventListenerInterface = (*EventListener)(nil)

func NewEventListener(
	consumer KafkaConsumerInterface,
	logger zerolog.Logger,
	topic string,
	maxInactivity time.Duration,
	kafkaPingTimeout time.Duration) *EventListener {

	return &EventListener{
		Consumer:         consumer,
		Logger:           logger,
		Topic:            topic,
		lastActivityTime: time.Now(),
		MaxInactivity:    maxInactivity,
		KafkaPingTimeout: kafkaPingTimeout,
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
	ticker := time.NewTicker(l.MaxInactivity / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.Logger.Info().Str("topic", l.Topic).Msg("Stopping event listener")
			return nil
		case <-ticker.C:
			l.mu.Lock()
			l.lastActivityTime = time.Now()
			l.mu.Unlock()
			l.Logger.Debug().Msg("EventListener heartbeat updated (no message received)")
		default:
			msg, err := l.Consumer.ReadMessage(100 * time.Millisecond)
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

			if len(baseEvent.EventType) >= 12 && baseEvent.EventType[:12] == "user.action." {
				var event events.UserActionEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					l.Logger.Error().Err(err).Str("event_type", baseEvent.EventType).Msg("Failed to unmarshal user action event")
					continue
				}
				err = l.HandleUserActionEvent(ctx, &event)
			} else {
				l.Logger.Warn().Str("event_type", baseEvent.EventType).Msg("Unknown event type received")
				continue
			}

			if err != nil {
				l.Logger.Error().Err(err).Str("event_type", baseEvent.EventType).Msg("Failed to handle event")
				continue
			}

			l.mu.Lock()
			l.lastActivityTime = time.Now()
			l.mu.Unlock()
			l.Logger.Debug().Msg("EventListener activity time updated (message processed)")

			_, err = l.Consumer.CommitMessage(msg)
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

func (l *EventListener) Ping() error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if time.Since(l.lastActivityTime) > l.MaxInactivity {
		l.Logger.Warn().
			Stringer("last_activity", l.lastActivityTime).
			Stringer("max_inactivity", l.MaxInactivity).
			Msg("Consumer listener loop inactive for too long")
		return errors.New("consumer listener loop inactive for too long")
	}

	timeoutMs := int(l.KafkaPingTimeout / time.Millisecond)

	_, err := l.Consumer.GetMetadata(&l.Topic, true, timeoutMs)
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && (kafkaErr.IsRetriable() || kafkaErr.Code() == kafka.ErrTimedOut) {
			l.Logger.Warn().Err(err).Msg("Kafka consumer connection check returned a retriable error. Still considered healthy for now.")
			return nil
		}
		l.Logger.Error().Err(err).Msg("Kafka consumer failed to get metadata during health check (connection issue)")
		return errors.New("kafka consumer not connected or unhealthy")
	}

	return nil
}
