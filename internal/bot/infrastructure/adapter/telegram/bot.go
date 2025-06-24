package telegram

import (
	"fmt"
	"log/slog"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/cache"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/inmemory"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/storage"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

type TgBot struct {
	Bot          *telebot.Bot
	logger       *slog.Logger
	repo         storage.Repo
	Scrapper     scrapper.ScraperClient
	stateManager domain.StateManager
	Cache        cache.Cache
}

func NewBot(
	cfg *config.Config,
	repo storage.Repo,
	scraper scrapper.ScraperClient,
	logger *slog.Logger,
	cacheRedis cache.Cache,
) (*TgBot, error) {
	pref := telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)

	if err != nil {
		return nil, fmt.Errorf("error when creating a bot: %s", err.Error())
	}

	tgBot := &TgBot{
		Bot:          bot,
		logger:       logger,
		repo:         repo,
		Scrapper:     scraper,
		stateManager: inmemory.NewStateManager(),
		Cache:        cacheRedis,
	}

	return tgBot, nil
}
