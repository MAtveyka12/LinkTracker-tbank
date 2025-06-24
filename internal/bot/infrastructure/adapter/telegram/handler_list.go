package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

func (t *TgBot) List(ctx telebot.Context) error {
	userID := ctx.Sender().ID
	key := fmt.Sprintf("links:%d", userID)

	if _, exists := domain.Users[userID]; !exists {
		return ctx.Send("Сначала зарегистрируйтесь с помощью /start.")
	}

	var links []string

	cached, err := t.Cache.Get(context.Background(), key)
	if err == nil && cached != "" {
		err = json.Unmarshal([]byte(cached), &links)

		if err != nil {
			t.logger.Error("Failed to unmarshal", slog.String("error", err.Error()))
			return err
		}

		t.logger.Info("Returned from the cache", slog.Int64("user_id", userID))
	} else {
		links, err = t.getLinksFromScraperAndCache(ctx, userID, key)
		if err != nil {
			return err
		}
	}

	if len(links) == 0 {
		return ctx.Send("У вас нет отслеживаемых ссылок.")
	}

	message := "Вы отслеживаете ссылки:\n"
	for _, link := range links {
		message += fmt.Sprintf("- %s\n", link)
	}

	return ctx.Send(message)
}

func (t *TgBot) getLinksFromScraperAndCache(ctx telebot.Context, userID int64, key string) ([]string, error) {
	links, err := t.Scrapper.GetLinks(context.Background(), userID)
	if err != nil {
		if err := ctx.Send("Произошла ошибка при получении списка ссылок."); err != nil {
			slog.Error(fmt.Sprintf("couldn't send message: %s", err.Error()))
		}

		return nil, fmt.Errorf("error when getting a list of links from the scraper: %s", err.Error())
	}

	data, err := json.Marshal(links)
	if err != nil {
		t.logger.Error("Failed to marshal", slog.String("error", err.Error()))
		return nil, err
	}

	err = t.Cache.Set(context.Background(), key, string(data))
	if err != nil {
		t.logger.Error(
			"failed to save data in cache",
			slog.String("op", "Cache.Set"),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("cache set failed: %s", err.Error())
	}

	return links, nil
}

func (t *TgBot) ListHandler() {
	t.Bot.Handle("/list", t.List)
}
