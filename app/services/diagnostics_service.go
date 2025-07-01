package services

import (
	"users-service/app/repositories"

	"github.com/rs/zerolog"
)

type DiagnosticsService struct {
	UsersRepository repositories.UsersRepositoryInterface
	Logger          zerolog.Logger
}

type DiagnosticsServiceInterface interface {
}

var _ DiagnosticsServiceInterface = (*DiagnosticsService)(nil)

func NewDiagnosticsService(
	usersRepo repositories.UsersRepositoryInterface,
	logger zerolog.Logger) *DiagnosticsService {

	return &DiagnosticsService{
		UsersRepository: usersRepo,
		Logger:          logger,
	}
}
