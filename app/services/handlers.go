package services

import (
	"context"
	"errors"
	"users-service/app/models/ent"
)

func HandleUserServiceCall[R any](
	ctx context.Context,
	userID int,
	s UsersService,
	req R,
	logic func(ctx context.Context, user *ent.User, req R) (*ent.User, error),
) (*ent.User, error) {
	var err error
	var updatedUser *ent.User

	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updatedUser, err = logic(ctx, user, req)
	if err != nil {
		return nil, err
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	return updatedUser, nil
}

func HandleUserAction(
	ctx context.Context,
	userID int,
	s *UsersService,
	actionLogic func(ctx context.Context, user *ent.User) error,
) error {
	user, err := s.UsersRepository.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("failed to fetch user")
	}

	if err := actionLogic(ctx, user); err != nil {
		return err
	}

	if err := s.EventNotifier.CreateNotificationEvent(user.Edges.UserSettings, user.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to produce notification event")
		return errors.New("failed to send notification")
	}

	return nil
}
