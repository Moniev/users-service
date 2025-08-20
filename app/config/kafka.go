package config

import (
	"context"
	"fmt"
	"time"
	"users-service/app/models/handlers"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

func CreateTopics(logger zerolog.Logger, settings *Settings, topics []string) error {
	config := &kafka.ConfigMap{
		"bootstrap.servers":                     settings.KafkaBootstrapServers,
		"request.timeout.ms":                    30000,
		"security.protocol":                     settings.KafkaSecurityProtocol,
		"ssl.ca.location":                       settings.KafkaSslCaPath,
		"ssl.endpoint.identification.algorithm": "https",
	}

	if settings.KafkaSslCertPath != "" && settings.KafkaSslKeyPath != "" {
		logger.Info().Msg("Found certificate and client key.")
		if err := config.SetKey("ssl.certificate.location", settings.KafkaSslCertPath); err != nil {
			return err
		}

		if err := config.SetKey("ssl.key.location", settings.KafkaSslKeyPath); err != nil {
			return err
		}

		if settings.KafkaClientKeyPassword != "" {
			logger.Info().Msg("Found client's key")
			if err := config.SetKey("ssl.key.password", settings.KafkaClientKeyPassword); err != nil {
				return err
			}
		}
	}

	adminClient, err := kafka.NewAdminClient(config)
	if err != nil {
		return fmt.Errorf("failed to create kafka admin klient: %w", err)
	}
	defer adminClient.Close()

	var topicSpecs []kafka.TopicSpecification
	for _, topic := range topics {
		topicSpecs = append(topicSpecs, kafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     10,
			ReplicationFactor: 1,
		})
	}

	if len(topicSpecs) == 0 {
		logger.Info().Msg("No Kafka topics specified for creation, skipping.")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := adminClient.CreateTopics(ctx, topicSpecs)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to issue CreateTopics request")
		return fmt.Errorf("failed to issue create topics request: %w", err)
	}

	var creationError error
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			logger.Error().Err(result.Error).Str("topic", result.Topic).Msg("Failed to create topic")
			if creationError == nil {
				creationError = result.Error
			}
		} else if result.Error.Code() == kafka.ErrTopicAlreadyExists {
			logger.Info().Str("topic", result.Topic).Msg("Topic already exists, skipping.")
		} else {
			logger.Info().Str("topic", result.Topic).Msg("Topic created successfully.")
		}
	}

	return creationError
}

func NewKafkaConsumer(
	groupID,
	topic string,
	handler handlers.MessageHandler,
	logger zerolog.Logger,
	settings *Settings) (*handlers.ConsumerWrapper, error) {

	config := &kafka.ConfigMap{
		"bootstrap.servers":                     settings.KafkaBootstrapServers,
		"security.protocol":                     settings.KafkaSecurityProtocol,
		"ssl.ca.location":                       settings.KafkaSslCaPath,
		"ssl.endpoint.identification.algorithm": "https",
		"group.id":                              groupID,
		"auto.offset.reset":                     "earliest",
		"enable.auto.commit":                    false,
	}

	if settings.KafkaSslCertPath != "" && settings.KafkaSslKeyPath != "" {
		logger.Info().Msg("Client certificate and key found. Applying mTLS configuration.")
		if err := config.SetKey("ssl.certificate.location", settings.KafkaSslCertPath); err != nil {
			return nil, fmt.Errorf("failed to set ssl.certificate.location: %w", err)
		}

		if err := config.SetKey("ssl.key.location", settings.KafkaSslKeyPath); err != nil {
			return nil, fmt.Errorf("failed to set ssl.key.location: %w", err)
		}

		if settings.KafkaClientKeyPassword != "" {
			logger.Info().Msg("Found client's key password.")
			if err := config.SetKey("ssl.key.password", settings.KafkaClientKeyPassword); err != nil {
				return nil, fmt.Errorf("failed to set ssl.key.password: %w", err)
			}
		}
	} else {
		logger.Warn().Msg("Client certificate or key not found. mTLS may fail if required by broker.")
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Kafka consumer")
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	rebalanceCb := func(c *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			logger.Info().
				Str("partitions", fmt.Sprintf("%v", e.Partitions)).
				Msg("Partitions assigned")
			c.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			committedOffsets, err := c.Commit()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to commit offsets on revoke")
			} else {
				logger.Info().
					Str("offsets", fmt.Sprintf("%v", committedOffsets)).
					Msg("Offsets committed on partition revocation")
			}
		}
		return nil
	}

	err = consumer.Subscribe(topic, rebalanceCb)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to subscribe to Kafka topic")
		consumer.Close()
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	logger.Info().Str("topic", topic).Str("groupID", groupID).Msg("Kafka consumer initialized and subscribed successfully")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer consumer.Close()

		for {
			select {
			case <-ctx.Done():
				logger.Info().Str("topic", topic).Msg("Stopping consumer loop.")
				return
			default:
				ev := consumer.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					if handler != nil {
						if err := handler.Handle(ctx, e); err != nil {
							logger.Error().Err(err).Str("topic", *e.TopicPartition.Topic).Msg("Failed to process message")
							continue
						}
					}
					_, err := consumer.CommitMessage(e)
					if err != nil {
						logger.Error().Err(err).Msg("Failed to commit offset")
					} else {
						logger.Debug().
							Str("topic", *e.TopicPartition.Topic).
							Int32("partition", e.TopicPartition.Partition).
							Int64("offset", int64(e.TopicPartition.Offset)).
							Msg("Offset committed")
					}
				case kafka.Error:
					logger.Error().Err(e).Msg("Consumer error")
				default:
					logger.Debug().Str("event", fmt.Sprintf("%v", e)).Msg("Ignored event")
				}
			}
		}
	}()

	return &handlers.ConsumerWrapper{
		Consumer: consumer,
		Cancel:   cancel,
	}, nil
}

func NewKafkaProducer(logger zerolog.Logger, settings *Settings) (*kafka.Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":                     settings.KafkaBootstrapServers,
		"acks":                                  "all",
		"ssl.endpoint.identification.algorithm": "https",
		"linger.ms":                             5,
		"queue.buffering.max.ms":                5000,
		"queue.buffering.max.kbytes":            1048576,
		"batch.num.messages":                    10000,
		"compression.type":                      "gzip",
		"retries":                               5,
		"retry.backoff.ms":                      100,
		"partitioner":                           "murmur2_random",
		"enable.idempotence":                    true,
	}

	if settings.KafkaSecurityProtocol != "" {
		if err := config.SetKey("security.protocol", settings.KafkaSecurityProtocol); err != nil {
			return nil, err
		}
	}

	if settings.KafkaSslCaPath != "" {
		if err := config.SetKey("ssl.ca.location", settings.KafkaSslCaPath); err != nil {
			return nil, err
		}
	}

	clientCertPath := settings.KafkaSslCertPath
	clientKeyPath := settings.KafkaSslKeyPath

	if clientCertPath != "" && clientKeyPath != "" {
		logger.Info().Msg("Client certificate and key found. Applying mTLS configuration.")
		if err := config.SetKey("ssl.certificate.location", clientCertPath); err != nil {
			return nil, err
		}
		if err := config.SetKey("ssl.key.location", clientKeyPath); err != nil {
			return nil, err
		}
	} else {
		logger.Info().Msg("Client certificate or key not found. Using one-way TLS.")
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Kafka producer")
		return nil, err
	}

	go func() {
		for event := range producer.Events() {
			switch eventTyped := event.(type) {
			case *kafka.Message:
				if eventTyped.TopicPartition.Error != nil {
					logger.Error().
						Err(eventTyped.TopicPartition.Error).
						Str("topic", *eventTyped.TopicPartition.Topic).
						Int32("partition", eventTyped.TopicPartition.Partition).
						Msg("Failed to deliver message")
				} else {
					logger.Debug().
						Str("topic", *eventTyped.TopicPartition.Topic).
						Int32("partition", eventTyped.TopicPartition.Partition).
						Msg("Message delivered")
				}
			case kafka.Error:
				logger.Error().Err(eventTyped).Msg("Producer error")
			}
		}
	}()

	logger.Info().Str("broker_id", settings.KafkaBrokerID).Msg("Kafka producer initialized successfully")
	return producer, nil
}
