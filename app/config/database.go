package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"users-service/app/middlewares"
	"users-service/app/models/ent"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

func NewEntClient(ctx context.Context, logger zerolog.Logger, settings *Settings) (*ent.Client, *sql.Driver, error) {
	encodedUser := url.QueryEscape(settings.DBUser)
	encodedPassword := url.QueryEscape(settings.DBPassword)
	encodedDBName := url.QueryEscape(settings.DBName)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=verify-full",
		encodedUser, encodedPassword, settings.DBHost, settings.DBPort, encodedDBName)

	caCert, err := os.ReadFile(settings.DbSslCaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load cert path: %s: %w", settings.DbSslCaPath, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(settings.DbSslCertPath, settings.DbSslKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load keys pair for database: %w", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
		ServerName:   settings.DBHost,
	}

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	config.TLSConfig = tlsConfig

	db := stdlib.OpenDB(*config)
	driver := sql.OpenDB(dialect.Postgres, db)

	client := ent.NewClient(ent.Driver(driver))
	client.Use(middlewares.UpdateDBMetrics())

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := client.Schema.Create(ctx); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create database: %w", err)
	}

	logger.Info().Msg("Initialized ent client successfully")
	return client, driver, nil
}
