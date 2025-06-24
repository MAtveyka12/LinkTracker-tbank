package telegram

import (
	"context"
	"log/slog"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

func (t *TgBot) Message(ctx telebot.Context) error {
	userID := ctx.Sender().ID

	if data, exists := domain.UserRegistrationData[userID]; exists {
		switch data.State {
		case domain.StateWaitingForEmail:
			t.logger.Info("Processing email registration step")

			return t.HandleRegistrationEmail(ctx, userID, data)
		case domain.StateWaitingForPassword:
			t.logger.Info("Processing password registration step")

			return t.HandleRegistrationPassword(ctx, userID, data)
		default:
			t.logger.Warn("Unknown registration state",
				slog.Int("state", data.State),
			)

			return nil
		}
	}

	state, exists, err := t.stateManager.GetTrackState(context.Background(), userID)
	if err != nil {
		t.logger.Error("failed to get tracking state", slog.String("error", err.Error()))
		return err
	}

	if exists && state == domain.StateWaitingForLink {
		return t.HandleTrackLink(ctx, userID)
	}

	untrackState, exists, err := t.stateManager.GetUntrackState(context.Background(), userID)
	if err != nil {
		t.logger.Error("failed to get untracking state", slog.String("error", err.Error()))
		return err
	}

	if exists && untrackState {
		return t.HandleUntrackLink(ctx, userID)
	}

	return t.HandleUnknownCommand(ctx)
}

func (t *TgBot) MessageHandler() {
	t.Bot.Handle(telebot.OnText, t.Message)
}
