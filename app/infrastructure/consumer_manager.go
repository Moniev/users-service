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
	return &ConsumerManager{
		ActiveConsumers: make(map[string]handlers.ConsumerInterface),
		Logger:          logger,
	}
}

func (m *ConsumerManager) Register(topic string, consumer handlers.ConsumerInterface) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	if consumer == nil {
		return fmt.Errorf("consumer cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.ActiveConsumers[topic]; exists {
		m.Logger.Warn().Str("topic", topic).Msg("overwriting existing consumer for topic")
	}

	m.ActiveConsumers[topic] = consumer
	m.Logger.Info().Str("topic", topic).Msg("consumer registered successfully")
	return nil
}

func (m *ConsumerManager) PingAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for topic, consumer := range m.ActiveConsumers {
		_, err := consumer.GetMetadata(nil, true, 1000)
		if err != nil {
			m.Logger.Error().
				Str("topic", topic).
				Err(err).
				Msg("consumer health check failed")
			return fmt.Errorf("consumer for topic '%s' is unhealthy: %w", topic, err)
		}
		m.Logger.Debug().Str("topic", topic).Msg("consumer health check passed")
	}

	return nil
}

func (m *ConsumerManager) GetConsumer(topic string) (handlers.ConsumerInterface, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	consumer, exists := m.ActiveConsumers[topic]
	return consumer, exists
}
