package config

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/spf13/viper"
)

var (
	config Settings
	once   sync.Once
)

type Settings struct {
	ApiVersion  string `mapstructure:"API_VERSION"`
	ApiPort     string `mapstructure:"API_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	TestMode    string `mapstructure:"TEST_MODE"`

	EncryptionSecretKey string `mapstructure:"ENCRYPTION_SECRET_KEY"`
	JwtPrivateKeyPath   string `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	JwtPublicKeyPath    string `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	JwtPrivateKeyBase64 string `mapstructure:"JWT_PRIVATE_KEY_BASE64"`
	JwtPublicKeyBase64  string `mapstructure:"JWT_PUBLIC_KEY_BASE64"`

	DBUser        string `mapstructure:"DB_USER"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBHost        string `mapstructure:"DB_HOST"`
	DBPort        int    `mapstructure:"DB_PORT"`
	DBName        string `mapstructure:"DB_NAME"`
	DbSslCaPath   string `mapstructure:"DB_SSL_CA_PATH"`
	DbSslCertPath string `mapstructure:"DB_SSL_CERT_PATH"`
	DbSslKeyPath  string `mapstructure:"DB_SSL_KEY_PATH"`

	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        int    `mapstructure:"REDIS_PORT"`
	RedisDB          int    `mapstructure:"REDIS_DB"`
	RedisPassword    string `mapstructure:"REDIS_PASSWORD"`
	RedisTlsEnabled  bool   `mapstructure:"REDIS_TLS_ENABLED"`
	RedisSslCaPath   string `mapstructure:"REDIS_SSL_CA_PATH"`
	RedisSslCertPath string `mapstructure:"REDIS_SSL_CERT_PATH"`
	RedisSslKeyPath  string `mapstructure:"REDIS_SSL_KEY_PATH"`

	KafkaBootstrapServers  string   `mapstructure:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaClientID          string   `mapstructure:"KAFKA_CLIENT_ID"`
	KafkaSecurityProtocol  string   `mapstructure:"KAFKA_SECURITY_PROTOCOL"`
	KafkaSaslMechanism     string   `mapstructure:"KAFKA_SASL_MECHANISM"`
	KafkaSaslUsername      string   `mapstructure:"KAFKA_SASL_USERNAME"`
	KafkaSaslPassword      string   `mapstructure:"KAFKA_SASL_PASSWORD"`
	KafkaSslCaPath         string   `mapstructure:"KAFKA_SSL_CA_PATH"`
	KafkaSslCertPath       string   `mapstructure:"KAFKA_SSL_CERT_PATH"`
	KafkaSslKeyPath        string   `mapstructure:"KAFKA_SSL_KEY_PATH"`
	KafkaClientKeyPassword string   `mapstructure:"KAFKA_CLIENT_KEY_PASSWORD"`
	KafkaTopics            []string `mapstructure:"KAFKA_TOPICS"`
	KafkaBrokerID          string   `mapstructure:"KAFKA_BROKER_ID"`
	KafkaNotifierTopic     string   `mapstructure:"KAFKA_NOTFIER_TOPIC"`

	GoogleApplicationCredentialsPath string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS_PATH"`
	GoogleClientID                   string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret               string `mapstructure:"GOOGLE_CLIENT_SECRET"`

	TwilioID       string `mapstructure:"TWILIO_ID"`
	TwilioPassword string `mapstructure:"TWILIO_PASSWORD"`
	StripeAPIKey   string `mapstructure:"STRIPE_API_KEY"`
}

func (s *Settings) DatabaseURI() string {
	dsn := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(s.DBUser, s.DBPassword),
		Host:   fmt.Sprintf("%s:%d", s.DBHost, s.DBPort),
		Path:   s.DBName,
	}
	return dsn.String()
}

func (s *Settings) RedisURI() string {
	scheme := "redis"
	if s.RedisTlsEnabled {
		scheme = "rediss"
	}

	auth := ""
	if s.RedisPassword != "" {
		auth = ":" + s.RedisPassword + "@"
	}

	return fmt.Sprintf("%s://%s%s:%d/%d", scheme, auth, s.RedisHost, s.RedisPort, s.RedisDB)
}

func GetSettings() *Settings {
	once.Do(func() {
		v := viper.New()

		v.SetConfigFile(".env")
		v.SetConfigType("env")
		v.AutomaticEnv()

		v.SetDefault("API_PORT", ":8000")
		v.SetDefault("ENVIRONMENT", "development")
		v.SetDefault("LOG_LEVEL", "info")
		v.SetDefault("ENCRYPTION_SECRET_KEY", "d644069bf1be3d82d432925ae9a9c4c2c94ad0dd27a65c9abb81482586c9404a")
		v.SetDefault("DB_HOST", "postgres.postgres.svc.cluster.local")
		v.SetDefault("DB_PORT", 5432)
		v.SetDefault("KAFKA_BOOTSTRAP_SERVERS", "kafka.kafka.svc.cluster.local:9093")
		v.SetDefault("KAFKA_CLIENT_ID", "users-service")
		v.SetDefault("KAFKA_SECURITY_PROTOCOL", "SSL")
		v.SetDefault("KAFKA_SSL_CA_PATH", "/etc/users-service-tls/kafka/ca.crt")
		v.SetDefault("KAFKA_SSL_CERT_PATH", "/etc/users-service-tls/kafka/tls.crt")
		v.SetDefault("KAFKA_SSL_KEY_PATH", "/etc/users-service-tls/kafka/tls.key")
		v.SetDefault("KAFKA_NOTIFIER_TOPIC", "notifications")
		v.SetDefault("REDIS_HOST", "redis.redis.svc.cluster.local")
		v.SetDefault("REDIS_PORT", 6379)
		v.SetDefault("REDIS_DB", 0)
		v.SetDefault("REDIS_TLS_ENABLED", false)
		v.SetDefault("REDIS_SSL_CA_PATH", "/etc/users-service-tls/redis/ca.crt")
		v.SetDefault("REDIS_SSL_CERT_PATH", "/etc/users-service-tls/redis/tls.crt")
		v.SetDefault("REDIS_SSL_KEY_PATH", "/etc/users-service-tls/redis/tls.key")

		v.SetDefault("TEST_MODE", "")
		v.SetDefault("JWT_PRIVATE_KEY_PATH", "")
		v.SetDefault("JWT_PUBLIC_KEY_PATH", "")
		v.SetDefault("DB_USER", "")
		v.SetDefault("DB_PASSWORD", "")
		v.SetDefault("DB_SSL_CA_PATH", "")
		v.SetDefault("DB_SSL_CERT_PATH", "")
		v.SetDefault("DB_SSL_KEY_PATH", "")
		v.SetDefault("REDIS_PASSWORD", "")
		v.SetDefault("KAFKA_CLIENT_KEY_PASSWORD", "")
		v.SetDefault("GOOGLE_APPLICATION_CREDENTIALS_PATH", "")
		v.SetDefault("GOOGLE_CLIENT_ID", "")
		v.SetDefault("GOOGLE_CLIENT_SECRET", "")
		v.SetDefault("TWILIO_ID", "")
		v.SetDefault("TWILIO_PASSWORD", "")
		v.SetDefault("STRIPE_API_KEY", "")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Printf("Warning: failed to load .env: %v", err)
			}
		}

		if err := v.Unmarshal(&config); err != nil {
			log.Fatalf("Failed to unmarshal config data: %v", err)
		}
	})

	return &config
}

func LoadJWTSigningKeys(settings *Settings) (crypto.PrivateKey, crypto.PublicKey, error) {
	if settings.JwtPrivateKeyBase64 != "" && settings.JwtPublicKeyBase64 != "" {
		return loadKeysFromBase64(settings.JwtPrivateKeyBase64, settings.JwtPublicKeyBase64)
	}

	return nil, nil, errors.New("no JWT key configuration provided (neither Base64 secrets nor file paths)")
}

func loadKeysFromBase64(privateKeyB64, publicKeyB64 string) (crypto.PrivateKey, crypto.PublicKey, error) {
	privatePEM, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to base64 decode private key: %w", err)
	}

	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return nil, nil, errors.New("failed to decode PEM block from base64 private key")
	}
	if privateBlock.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("unexpected type for private key PEM block: %s", privateBlock.Type)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse PKCS8 private key from base64: %w", err)
	}

	ed25519PrivateKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, errors.New("private key from base64 is not of type Ed25519")
	}

	publicPEM, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to base64 decode public key: %w", err)
	}

	publicBlock, _ := pem.Decode(publicPEM)
	if publicBlock == nil {
		return nil, nil, errors.New("failed to decode PEM block from base64 public key")
	}
	if publicBlock.Type != "PUBLIC KEY" {
		return nil, nil, fmt.Errorf("unexpected type for public key PEM block: %s", publicBlock.Type)
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse PKIX public key from base64: %w", err)
	}

	ed25519PublicKey, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return nil, nil, errors.New("public key from base64 is not of type Ed25519")
	}

	return ed25519PrivateKey, ed25519PublicKey, nil
}
