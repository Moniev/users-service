package tests

import (
	"context"
	"crypto/ed25519"
	"flag"
	"log"
	"os"
	"testing"
	"time"
	"users-service/app/config"
	"users-service/app/models/ent"
	"users-service/tests/registry"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/docker/docker/client"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestApp *registry.AppContainer
	TestDB  *ent.Client
)

var TestType = flag.String("test_type", "", "type of test to run (unit, integration, e2e)")

func TestMain(m *testing.M) {
	flag.Parse()

	switch *TestType {
	case "integration", "e2e":
		setupIntegrationEnvironment()
	case "unit":
		log.Println("--- Running in UNIT test mode. Skipping container setup. ---")
	default:
		log.Println("No -test_type flag provided, defaulting to 'unit'.")
		*TestType = "unit"
	}

	code := m.Run()
	os.Exit(code)
}

func setupIntegrationEnvironment() {
	ctx := context.Background()

	v := viper.New()
	v.SetConfigFile("settings/settings.yaml")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read test settings file: %s", err)
	}
	var testSettings *config.Settings
	if err := v.Unmarshal(&testSettings); err != nil {
		log.Fatalf("Failed to unmarshal test settings: %s", err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if level, err := zerolog.ParseLevel(testSettings.LogLevel); err == nil {
		logger = logger.Level(level)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(testSettings.DockerSocket))
	if err != nil {
		log.Printf("Failed to connect to docker daemon %s. Skipping integration tests", err)
		os.Exit(0)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		log.Printf("Docker daemon is turned off: %s. Skipping integration tests", err)
		os.Exit(0)
	}

	log.Println("--- Setting up INTEGRATION/E2E environment... ---")

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %s", err)
	}

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %s", err)
	}

	dbClient, err := ent.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("failed to connect to test postgres: %s", err)
	}

	if err := dbClient.Schema.Create(ctx); err != nil {
		log.Fatalf("failed to create schema: %s", err)
	}

	TestDB = dbClient
	redisContainer, err := tcredis.Run(ctx,
		"redis:7-alpine",
	)
	if err != nil {
		log.Fatalf("failed to start redis container: %s", err)
	}

	redisURI, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get redis connection string: %s", err)
	}

	opts, _ := redis.ParseURL(redisURI)
	redisClient := redis.NewClient(opts)

	kafkaContainer, err := tckafka.Run(ctx,
		"confluentinc/cp-kafka:7.3.2",
		tckafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		log.Fatalf("failed to start kafka container: %s", err)
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		log.Fatalf("failed to get kafka brokers: %s", err)
	}

	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers[0]})
	if err != nil {
		log.Fatalf("failed to create kafka producer: %s", err)
	}

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("failed to generate jwt keys: %s", err)
	}

	TestApp = registry.NewTestAppContainer(testSettings, dbClient, redisClient, kafkaProducer, logger, pubKey, privKey)

	log.Println("--- Test environment successfully initialized ---")
}
