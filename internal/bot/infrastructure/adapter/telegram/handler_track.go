package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
	"github.com/es-debug/backend-academy-2024-go-template/patterns"
)

var (
	githubRegex        *regexp.Regexp
	stackOverflowRegex *regexp.Regexp
)

func LoadPatterns() error {
	patternsPath := os.Getenv("PATTERNS_FILE_PATH")
	if patternsPath == "" {
		return fmt.Errorf("PATTERNS_FILE_PATH environment variable not set")
	}

	cfg, err := patterns.LoadPatterns(patternsPath)
	if err != nil {
		return fmt.Errorf("patterns not uploaded: %s", err.Error())
	}

	githubRegex = regexp.MustCompile(cfg.Patterns.GitHub)
	stackOverflowRegex = regexp.MustCompile(cfg.Patterns.StackOverflow)

	return nil
}

func (t *TgBot) Track(ctx telebot.Context) error {
	userID := ctx.Sender().ID

	if _, exists := domain.Users[userID]; !exists {
		return ctx.Send("Сначала зарегистрируйтесь с помощью /start.")
	}

	err := t.stateManager.SetTrackState(context.Background(), userID, domain.StateWaitingForLink)
	if err != nil {
		t.logger.Error("failed to set tracking state", slog.String("error", err.Error()))
		return err
	}

	t.logger.Info(fmt.Sprintf("The link will be added for the user %d", userID))

	err = ctx.Send("Введите ссылку для отслеживания:")
	if err != nil {
		return fmt.Errorf("incorrect data: %s", err.Error())
	}

	return nil
}

func (t *TgBot) HandleTrackLink(ctx telebot.Context, userID int64) error {
	link := ctx.Text()

	err := LoadPatterns()
	if err != nil {
		t.logger.Error("failed to load patterns", slog.String("error", err.Error()))
		return fmt.Errorf("configuration loading error: %s", err.Error())
	}

	if !githubRegex.MatchString(link) && !stackOverflowRegex.MatchString(link) {
		return ctx.Send("Некорректный формат ссылки. Поддерживаются только ссылки на GitHub репозитории или вопросы Stack Overflow.")
	}

	userLinks := t.repo.GetLinks(userID)
	for _, existingLink := range userLinks {
		if link == existingLink.URL {
			return ctx.Send("Эта ссылка уже отслеживается.")
		}
	}

	domain.UserTrackData[userID] = domain.Link{
		URL:    link,
		UserID: userID,
	}

	err = t.Scrapper.AddLink(context.Background(), domain.UserTrackData[userID])
	if err != nil {
		if err := ctx.Send("Произошла ошибка при добавлении ссылки на скрапер. Пожалуйста, попробуйте снова."); err != nil {
			slog.Error(fmt.Sprintf("couldn't send message: %s", err.Error()))
		}

		return fmt.Errorf("error when adding a link to the scrapper: %s", err.Error())
	}

	t.repo.AddLink(domain.UserTrackData[userID])

	if err := t.Cache.Delete(context.Background(), fmt.Sprintf("links:%d", userID)); err != nil {
		t.logger.Error("Failed to delete cache after track", slog.String("error", err.Error()))
	}

	url := domain.UserTrackData[userID].URL
	message := fmt.Sprintf("Ссылка %s добавлена.", url)

	err = t.stateManager.DeleteTrackState(context.Background(), userID)
	if err != nil {
		t.logger.Error("failed to delete tracking state", slog.String("error", err.Error()))
		return err
	}

	delete(domain.UserTrackData, userID)

	return ctx.Send(message)
}

func (t *TgBot) TrackHandler() {
	t.Bot.Handle("/track", t.Track)
}
