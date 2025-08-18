package services

import (
	"context"
	"errors"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	"users-service/app/models/requests"
	"users-service/app/repositories"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type UsersService struct {
	UsersRepository repositories.UsersRepositoryInterface
	EventNotifier   infrastructure.EventNotifierInterface
	Logger          zerolog.Logger
}

type UsersServiceInterface interface {
	Index(ctx *gin.Context, lim int) ([]*ent.User, error)

	UpdateUser(ctx context.Context, userID int, req *requests.User) (*ent.User, error)
	UpdateDetails(ctx context.Context, userID int, req *requests.Details) (*ent.User, error)
	UpdateSettings(ctx context.Context, userID int, req *requests.Settings) (*ent.User, error)
	UpdateEntrepreneurDetails(ctx context.Context, userID int, req *requests.EntrepreneurDetails) (*ent.User, error)
	UpdateLocation(ctx context.Context, userID int, req *requests.Location) (*ent.User, error)

	RemoveAccount(ctx context.Context, userID int) error

	GetUserPublic(ctx context.Context, userID int) (*ent.User, error)
	GetUserPrivate(ctx context.Context, userID int) (*ent.User, error)
}

var _ UsersServiceInterface = (*UsersService)(nil)

func NewUsersService(
	usersRepository repositories.UsersRepositoryInterface,
	eventNotifier infrastructure.EventNotifierInterface,
	logger zerolog.Logger) *UsersService {

	return &UsersService{
		UsersRepository: usersRepository,
		EventNotifier:   eventNotifier,
		Logger:          logger,
	}
}

func (s *UsersService) UpdateUser(ctx context.Context, userID int, req *requests.User) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	updatedUser, err := s.UsersRepository.UpdateUser(ctx, user, req)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	if updatedUser.Phone != user.Phone {
		if err := s.EventNotifier.CreateVerificationEvent(updatedUser.Edges.UserSettings, updatedUser.Phone, updatedUser.Edges.VerificationCode); err != nil {
			s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce verification event")
			return nil, errors.New("failed to send second factor code")
		}
	}

	if updatedUser.Mail != user.Mail {
		if err := s.EventNotifier.CreateRegistrationEvent(updatedUser, updatedUser.Edges.ActivationCode); err != nil {
			s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce activation event")
			return nil, errors.New("failed to send second factor code")
		}
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return updatedUser, nil
}

func (s *UsersService) UpdateDetails(ctx context.Context, userID int, req *requests.Details) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	updatedUser, err := s.UsersRepository.UpdateUsersDetails(ctx, user, req)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return updatedUser, nil
}

func (s *UsersService) UpdateSettings(ctx context.Context, userID int, req *requests.Settings) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	updatedUser, err := s.UsersRepository.UpdateUsersSettings(ctx, user, req)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return updatedUser, nil
}

func (s *UsersService) RemoveAccount(ctx context.Context, userID int) error {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("failed to fetch user")
	}

	if err := s.UsersRepository.RemoveAccount(ctx, user); err != nil {
		return errors.New("failed to remove user")
	}

	if err := s.EventNotifier.CreateNotificationEvent(user.Edges.UserSettings, user.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return errors.New("failed to send notification")
	}

	return nil
}

func (s *UsersService) UpdateEntrepreneurDetails(ctx context.Context, userID int, req *requests.EntrepreneurDetails) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	updatedUser, err := s.UsersRepository.UpdateEntrepreneurDetails(ctx, user, req)
	if err != nil {
		return nil, errors.New("failed to update user details")
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return user, nil
}

func (s UsersService) UpdateLocation(ctx context.Context, userID int, req *requests.Location) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	updatedUser, err := s.UsersRepository.UpdateLocation(ctx, user, req)
	if err != nil {
		return nil, errors.New("failed to update user details")
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return user, nil
}

func (s *UsersService) Index(ctx *gin.Context, lim int) ([]*ent.User, error) {
	users, err := s.UsersRepository.GetUsersPublic(ctx, lim)
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}

	return users, nil
}

func (s *UsersService) GetUserPublic(ctx context.Context, userID int) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserPublicByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	return user, nil
}

func (s *UsersService) GetUserPrivate(ctx context.Context, userID int) (*ent.User, error) {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	return user, nil
}
