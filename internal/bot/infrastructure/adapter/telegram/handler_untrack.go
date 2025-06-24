package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

func (t *TgBot) Untrack(ctx telebot.Context) error {
	userID := ctx.Sender().ID

	if _, exists := domain.Users[userID]; !exists {
		return ctx.Send("Сначала зарегистрируйтесь с помощью /start.")
	}

	links := t.repo.GetLinks(userID)

	if len(links) == 0 {
		return ctx.Send("У вас нет отслеживаемых ссылок.")
	}

	err := t.stateManager.SetUntrackState(context.Background(), userID, true)
	if err != nil {
		t.logger.Error("failed to set untracking state", slog.String("error", err.Error()))
		return err
	}

	err = ctx.Send("Введите ссылку, которую не хотите отслеживать:")
	if err != nil {
		return fmt.Errorf("incorrect data %s", err.Error())
	}

	return nil
}

func (t *TgBot) HandleUntrackLink(ctx telebot.Context, userID int64) error {
	link := ctx.Text()

	_, exists, err := t.stateManager.GetUntrackState(context.Background(), userID)
	if err != nil {
		t.logger.Error("failed to get untracking state", slog.String("error", err.Error()))
		return err
	}

	if !exists {
		return ctx.Send("Начните процесс удаления ссылки с команды /untrack.")
	}

	userLinks := t.repo.GetLinks(userID)
	if len(userLinks) == 0 {
		return ctx.Send("У вас нет отслеживаемых ссылок.")
	}

	if t.repo.DeleteLink(link, userID) {
		err := t.Scrapper.RemoveLink(context.Background(), link, userID)
		if err != nil {
			if err := ctx.Send("Произошла ошибка при удалении ссылки. Пожалуйста, попробуйте снова."); err != nil {
				slog.Error(fmt.Sprintf("couldn't send message: %s", err.Error()))
			}

			return fmt.Errorf("error when deleting a link on a scrapper: %s", err.Error())
		}

		err = t.stateManager.DeleteUntrackState(context.Background(), userID)
		if err != nil {
			t.logger.Error("failed to delete untracking state", slog.String("error", err.Error()))
			return nil
		}

		return ctx.Send(fmt.Sprintf("Ссылка %s больше не отслеживается.", link))
	}

	if err := t.Cache.Delete(context.Background(), fmt.Sprintf("links:%d", userID)); err != nil {
		t.logger.Error("Failed to delete cache after untrack", slog.String("error", err.Error()))
	}

	err = t.stateManager.DeleteUntrackState(context.Background(), userID)
	if err != nil {
		t.logger.Error("failed to delete untracking state", slog.String("error", err.Error()))
		return err
	}

	return ctx.Send("Такой ссылки нет.")
}

func (t *TgBot) UntrackHandler() {
	t.Bot.Handle("/untrack", t.Untrack)
}
