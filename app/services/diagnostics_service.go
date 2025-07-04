package services

import (
	"context"
	"fmt"
	"users-service/app/infrastructure"
	"users-service/app/repositories"

	"github.com/rs/zerolog"
)

type DiagnosticsService struct {
	UsersRepository repositories.UsersRepositoryInterface
	CacheStore      infrastructure.CacheStoreInterface
	EventNotifier   infrastructure.EventNotifierInterface
	ConsumerManager infrastructure.ConsumerManagerInterface
	Logger          zerolog.Logger
}

type DiagnosticsServiceInterface interface {
	CheckHealth(ctx context.Context) error
	CheckReadiness(ctx context.Context) error
}

var _ DiagnosticsServiceInterface = (*DiagnosticsService)(nil)

func NewDiagnosticsService(
	usersRepo repositories.UsersRepositoryInterface,
	cacheStore infrastructure.CacheStoreInterface,
	eventNotifier infrastructure.EventNotifierInterface,
	consumerManager infrastructure.ConsumerManagerInterface,
	logger zerolog.Logger) *DiagnosticsService {

	return &DiagnosticsService{
		UsersRepository: usersRepo,
		CacheStore:      cacheStore,
		EventNotifier:   eventNotifier,
		ConsumerManager: consumerManager,
		Logger:          logger,
	}
}

func (s *DiagnosticsService) CheckHealth(ctx context.Context) error {
	s.Logger.Debug().Msg("Performing health check...")

	if err := s.UsersRepository.Ping(); err != nil {
		s.Logger.Error().Err(err).Msg("Health check failed: UsersRepository ping failed")
		return fmt.Errorf("database unhealthy: %w", err)
	}

	s.Logger.Debug().Msg("Health check passed.")
	return nil
}

func (s *DiagnosticsService) CheckReadiness(ctx context.Context) error {
	if err := s.UsersRepository.Ping(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: UsersRepository ping failed")
		return fmt.Errorf("users repository not ready: %w", err)
	}

	if _, err := s.CacheStore.Ping(ctx).Result(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: CacheStore ping failed")
		return fmt.Errorf("cache store not ready: %w", err)
	}

	if err := s.ConsumerManager.PingAll(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: EventNotifier ping failed")
		return fmt.Errorf("consumers not ready: %w", err)
	}

	if err := s.EventNotifier.Ping(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: EventNotifier ping failed")
		return fmt.Errorf("event notifier not ready: %w", err)
	}

	s.Logger.Debug().Msg("Readiness check passed.")
	return nil
}
