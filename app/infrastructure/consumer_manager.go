package infrastructure

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	"users-service/app/models/handlers"
)

type ConsumerManagerInterface interface {
	GetConsumer(topic string) (handlers.ConsumerInterface, bool)
	Register(topic string, consumer handlers.ConsumerInterface) error
	PingAll() error
}

type ConsumerManager struct {
	mu              sync.RWMutex
	ActiveConsumers map[string]handlers.ConsumerInterface
	Logger          zerolog.Logger
}

var _ ConsumerManagerInterface = (*ConsumerManager)(nil)

func NewConsumerManager(logger zerolog.Logger) *ConsumerManager {
	logger.Info().Msg("Initializing new ConsumerManager")
	return &ConsumerManager{
		ActiveConsumers: make(map[string]handlers.ConsumerInterface),
		Logger:          logger,
	}
}

func (m *ConsumerManager) Register(topic string, consumer handlers.ConsumerInterface) error {
	m.Logger.Debug().Str("topic", topic).Msg("Attempting to register consumer")

	if topic == "" {
		m.Logger.Error().Msg("Failed to register consumer: topic cannot be empty")
		return fmt.Errorf("topic cannot be empty")
	}

	if consumer == nil {
		m.Logger.Error().Str("topic", topic).Msg("Failed to register consumer: consumer cannot be nil")
		return fmt.Errorf("consumer cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.ActiveConsumers[topic]; exists {
		m.Logger.Warn().Str("topic", topic).Msg("Overwriting existing consumer for topic")
	}

	m.ActiveConsumers[topic] = consumer
	m.Logger.Info().Str("topic", topic).Msg("Consumer registered successfully")
	return nil
}

func (m *ConsumerManager) PingAll() error {
	m.Logger.Info().Msg("Performing health check on all active consumers")
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.ActiveConsumers) == 0 {
		m.Logger.Info().Msg("No active consumers to ping")
		return nil
	}

	for topic, consumer := range m.ActiveConsumers {
		m.Logger.Debug().Str("topic", topic).Msg("Pinging consumer for health check")
		_, err := consumer.GetMetadata(nil, true, 1000)
		if err != nil {
			m.Logger.Error().
				Str("topic", topic).
				Err(err).
				Msg("Consumer health check failed")
			return fmt.Errorf("consumer for topic '%s' is unhealthy: %w", topic, err)
		}
		m.Logger.Debug().Str("topic", topic).Msg("Consumer health check passed")
	}

	m.Logger.Info().Msg("All active consumers passed health checks")
	return nil
}

func (m *ConsumerManager) GetConsumer(topic string) (handlers.ConsumerInterface, bool) {
	m.Logger.Debug().Str("topic", topic).Msg("Attempting to get consumer")
	m.mu.RLock()
	defer m.mu.RUnlock()
	consumer, exists := m.ActiveConsumers[topic]
	if exists {
		m.Logger.Debug().Str("topic", topic).Msg("Consumer found")
	} else {
		m.Logger.Warn().Str("topic", topic).Msg("Consumer not found")
	}
	return consumer, exists
}
