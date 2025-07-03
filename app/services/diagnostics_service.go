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
	EventListener   infrastructure.EventListenerInterface
	EventNotifier   infrastructure.EventNotifierInterface
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
	eventListener infrastructure.EventListenerInterface,
	eventNotifier infrastructure.EventNotifierInterface,
	logger zerolog.Logger) *DiagnosticsService {

	return &DiagnosticsService{
		UsersRepository: usersRepo,
		CacheStore:      cacheStore,
		EventListener:   eventListener,
		EventNotifier:   eventNotifier,
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

	if err := s.EventListener.Ping(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: EventListener ping failed")
		return fmt.Errorf("event listener not ready: %w", err)
	}

	if err := s.EventNotifier.Ping(); err != nil {
		s.Logger.Error().Err(err).Msg("Readiness check failed: EventNotifier ping failed")
		return fmt.Errorf("event notifier not ready: %w", err)
	}

	s.Logger.Debug().Msg("Readiness check passed.")
	return nil
}
