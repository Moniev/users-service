package services

import (
	"context"
	"errors"
	"fmt"
	"users-service/app/models/ent"
)

func HandleUserServiceCall[R any](
	ctx context.Context,
	userID int,
	s UsersService,
	req R,
	logic func(ctx context.Context, user *ent.User, req R) (*ent.User, error),
) (*ent.User, error) {
	requestType := fmt.Sprintf("%T", req)
	s.Logger.Info().Int("user_id", userID).Str("request_type", requestType).Msg("Handling user update call")

	user, err := s.UsersRepository.GetUserFunctionalByID(ctx, userID)
	if err != nil {
		s.Logger.Debug().Err(err).Int("user_id", userID).Msg("Failed to get user in helper")
		return nil, err
	}

	updatedUser, err := logic(ctx, user, req)
	if err != nil {
		s.Logger.Error().Err(err).Int("user_id", userID).Str("request_type", requestType).Msg("Update logic failed")
		return nil, err
	}

	if err := s.EventNotifier.CreateNotificationEvent(updatedUser.Edges.UserSettings, updatedUser.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to produce notification event")
		return nil, errors.New("failed to send notification")
	}

	s.Logger.Info().Int("user_id", userID).Str("request_type", requestType).Msg("User update call completed successfully")
	return updatedUser, nil
}

func HandleUserAction(
	ctx context.Context,
	userID int,
	s *UsersService,
	actionLogic func(ctx context.Context, user *ent.User) error,
) error {
	s.Logger.Info().Int("user_id", userID).Msg("Handling user action call")
	user, err := s.UsersRepository.GetUserFunctionalByID(ctx, userID)
	if err != nil {
		s.Logger.Debug().Err(err).Int("user_id", userID).Msg("Failed to get user in action helper")
		return errors.New("failed to fetch user")
	}

	if err := actionLogic(ctx, user); err != nil {
		s.Logger.Error().Err(err).Int("user_id", userID).Msg("Action logic failed")
		return err
	}

	if err := s.EventNotifier.CreateNotificationEvent(user.Edges.UserSettings, user.Edges.UserDevices); err != nil {
		s.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to produce notification event")
		return errors.New("failed to send notification")
	}

	s.Logger.Info().Int("user_id", userID).Msg("User action call completed successfully")
	return nil
}
