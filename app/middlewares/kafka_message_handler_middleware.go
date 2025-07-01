package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

// MessageHandler is a function type that defines how to handle a Kafka message.
// It takes a context, a Kafka message, and a logger, and returns an error.
type MessageHandler func(ctx context.Context, msg *kafka.Message, logger zerolog.Logger) error

// Middleware is a function type that wraps a MessageHandler.
// It allows chaining multiple middleware layers for a given handler.
type KafkaMiddleware func(next MessageHandler) MessageHandler

// ChainMiddleware applies a series of middleware functions to a MessageHandler,
// chaining them in the order they are provided. The middlewares are applied from
// last to first, with the last middleware wrapping the handler.
func ChainMiddleware(handler MessageHandler, middlewares ...KafkaMiddleware) MessageHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// LoggingMiddleware is a middleware that logs the processing duration of a message.
// It logs the time taken to process the message from the moment the handler starts until it completes.
func LoggingMiddleware(next MessageHandler) MessageHandler {
	return func(ctx context.Context, msg *kafka.Message, logger zerolog.Logger) error {
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			logger.Debug().
				Str("topic", *msg.TopicPartition.Topic).
				Dur("durationMS", duration).
				Msg("Message processing duration")
		}()

		return next(ctx, msg, logger)
	}
}

// MetricsMiddleware is a middleware that tracks and logs the number of processed messages.
// It increments a counter each time a message is processed and logs the updated count.
func MetricsMiddleware() KafkaMiddleware {
	var processedMessages int64
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *kafka.Message, logger zerolog.Logger) error {
			atomic.AddInt64(&processedMessages, 1)
			logger.Debug().Int64("processed_messages", atomic.LoadInt64(&processedMessages)).Msg("Message processed")
			return next(ctx, msg, logger)
		}
	}
}

// RetryMiddleware is a middleware that retries processing a message a specified number of times
// with exponential backoff. It retries on failure up to the maximum specified retries
func RetryMiddleware(maxRetries int, backoff time.Duration) KafkaMiddleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *kafka.Message, logger zerolog.Logger) error {
			for attempt := 1; attempt <= maxRetries; attempt++ {
				err := next(ctx, msg, logger)
				if err == nil {
					return nil
				}

				logger.Warn().
					Int("attempt", attempt).
					Int("maxRetries", maxRetries).
					Err(err).
					Msg("Retrying message processing")

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
				}
			}
			return fmt.Errorf("failed to process message after %d attempts", maxRetries)
		}
	}
}

// ValidationMiddleware is a middleware that validates the content of the Kafka message.
// It ensures that the message contains a non-empty value and a valid JSON with a required "event" field.
func ValidationMiddleware() KafkaMiddleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *kafka.Message, logger zerolog.Logger) error {
			if msg == nil || msg.Value == nil {
				return fmt.Errorf("invalid message: nil or empty value")
			}

			var data map[string]interface{}
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				return fmt.Errorf("invalid message: not a valid JSON: %w", err)
			}

			if _, ok := data["event"]; !ok {
				return fmt.Errorf("invalid message: missing 'event' field")
			}

			return next(ctx, msg, logger)
		}
	}
}
