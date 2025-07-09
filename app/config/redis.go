package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func NewRedisClient(ctx context.Context, logger zerolog.Logger, settings *Settings) *redis.Client {
	if settings.RedisHost == "" {
		logger.Fatal().Msg("REDIS_HOST isn't set properly in configuration")
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", settings.RedisHost, settings.RedisPort),
		Password: settings.RedisPassword,
		DB:       settings.RedisDB,
	}

	if settings.RedisTlsEnabled {
		if settings.RedisSslCertPath == "" || settings.RedisSslKeyPath == "" {
			logger.Fatal().Msg("REDIS_SSL_CERT_PATH nad REDIS_SSL_KEY_PATH are not set properly")
		}

		caCert, err := os.ReadFile(settings.RedisSslCaPath)
		if err != nil {
			logger.Fatal().Err(err).Str("path", settings.RedisSslCaPath).Msg("Failed to read Redis CA certificate")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			logger.Fatal().Str("path", settings.RedisSslCaPath).Msg("Failed to parse Redis CA certificate")
		}

		clientCert, err := tls.LoadX509KeyPair(settings.RedisSslCertPath, settings.RedisSslKeyPath)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to load Redis SSL Certificate and SSL key")
		}

		opts.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{clientCert},
			ServerName:   settings.RedisHost,
		}
	}

	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Fatal().Err(err).Str("address", opts.Addr).Msg("Failed to connect to Redis server")
	}

	logger.Info().Str("address", opts.Addr).Int("db", opts.DB).Msg("Successfully connected to Redis server")
	return client
}
