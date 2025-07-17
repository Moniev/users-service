package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"users-service/app/models/events"
	"users-service/app/repositories"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type UserActionHandler struct {
	UsersRepository repositories.UsersRepositoryInterface
	Logger          zerolog.Logger
}

func NewUserActionHandler(
	usersRepository repositories.UsersRepositoryInterface,
	logger zerolog.Logger,
) *UserActionHandler {

	return &UserActionHandler{
		UsersRepository: usersRepository,
		Logger:          logger,
	}
}

type UserEventHandlerInterface interface {
	Handle(ctx context.Context, msg *kafka.Message) error
}

func (h *UserActionHandler) Handle(ctx context.Context, msg *kafka.Message) error {
	var baseEvent events.BaseEvent
	if err := json.Unmarshal(msg.Value, &baseEvent); err != nil {
		return fmt.Errorf("failed to unmarshal base event: %w", err)
	}

	if len(baseEvent.EventType) >= 12 && baseEvent.EventType[:12] == "user.action." {
		var event events.UserActionEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("failed to unmarshal user action event: %w", err)
		}

		switch event.Action {

		case "":
			return errors.New("invalid user action event: missing action")
		case "action":
			user, err := h.UsersRepository.GetUserByID(ctx, event.UserID)
			if err != nil {
				return err
			}

			return h.UsersRepository.CreateUserAction(ctx, user, event.Action, event.OriginDevice, event.Details)
		case "subscription":
		}

		h.Logger.Debug().
			Str("event_id", event.EventID).
			Int("user_id", event.UserID).
			Str("action", event.Action).
			Msg("Processed user action event")

		return nil
	}

	h.Logger.Warn().Str("event_type", baseEvent.EventType).Msg("Unknown event type received by UserActionHandler")
	return nil
}
