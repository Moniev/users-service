package utils

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ConsumerWrapper encapsulates a Kafka consumer with cancellation functionality.
// It provides a way to manage a Kafka consumer instance and gracefully stop its operation.
type ConsumerWrapper struct {
	Consumer *kafka.Consumer    // Kafka consumer instance for reading messages
	Cancel   context.CancelFunc // Function to cancel the consumer's operation
}
