package telegram_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/adapter/telegram"
	mockCache "github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/cache/mock"
	mockScrapper "github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/scrapper/mock"
)

func TestUnknownCommand(t *testing.T) {
	bot := &telegram.TgBot{}
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCtx := telegram.NewMockContext(ctrl)

	mockCtx.EXPECT().Send("Неизвестная команда. Введите /help для списка доступных команд.").Return(nil)

	err := bot.HandleUnknownCommand(mockCtx)
	assert.NoError(t, err)
}

func TestList_NoRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bot := &telegram.TgBot{}
	mockCtx := telegram.NewMockContext(ctrl)

	mockCtx.EXPECT().Sender().Return(&telebot.User{ID: 123})
	mockCtx.EXPECT().Send("Сначала зарегистрируйтесь с помощью /start.").Return(nil)

	domain.Users = make(map[int64]domain.User)

	err := bot.List(mockCtx)
	assert.NoError(t, err)
}

func TestList_NoLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	scraperMock := mockScrapper.NewMockScraperClient(ctrl)
	cacheMock := mockCache.NewMockCache(ctrl)
	mockCtx := telegram.NewMockContext(ctrl)

	userID := int64(123)
	key := fmt.Sprintf("links:%d", userID)

	mockCtx.EXPECT().Sender().Return(&telebot.User{ID: userID})
	cacheMock.EXPECT().Get(gomock.Any(), key).Return("", fmt.Errorf("not found"))
	scraperMock.EXPECT().GetLinks(context.Background(), userID).Return([]string{}, nil)
	cacheMock.EXPECT().Set(gomock.Any(), key, "[]").Return(nil)
	mockCtx.EXPECT().Send("У вас нет отслеживаемых ссылок.").Return(nil)

	domain.Users = map[int64]domain.User{userID: {}}

	bot := &telegram.TgBot{
		Scrapper: scraperMock,
		Cache:    cacheMock,
	}

	err := bot.List(mockCtx)
	assert.NoError(t, err)
}

func TestList_WithLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	scraperMock := mockScrapper.NewMockScraperClient(ctrl)
	cacheMock := mockCache.NewMockCache(ctrl)
	mockCtx := telegram.NewMockContext(ctrl)

	bot := &telegram.TgBot{
		Scrapper: scraperMock,
		Cache:    cacheMock,
	}

	userID := int64(123)
	key := fmt.Sprintf("links:%d", userID)
	links := []string{"https://example.com", "https://golang.org"}
	linksJSON, _ := json.Marshal(links)

	mockCtx.EXPECT().Sender().Return(&telebot.User{ID: userID})
	cacheMock.EXPECT().Get(gomock.Any(), key).Return("", fmt.Errorf("not found"))
	scraperMock.EXPECT().GetLinks(context.Background(), userID).Return(links, nil)
	cacheMock.EXPECT().Set(gomock.Any(), key, string(linksJSON)).Return(nil)
	mockCtx.EXPECT().Send("Вы отслеживаете ссылки:\n- https://example.com\n- https://golang.org\n").Return(nil)

	domain.Users = map[int64]domain.User{userID: {}}

	err := bot.List(mockCtx)
	assert.NoError(t, err)
}

func TestList_ErrorGettingLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	scraperMock := mockScrapper.NewMockScraperClient(ctrl)
	cacheMock := mockCache.NewMockCache(ctrl)
	mockCtx := telegram.NewMockContext(ctrl)

	bot := &telegram.TgBot{
		Scrapper: scraperMock,
		Cache:    cacheMock,
	}

	userID := int64(123)
	key := fmt.Sprintf("links:%d", userID)

	mockCtx.EXPECT().Sender().Return(&telebot.User{ID: userID})
	cacheMock.EXPECT().Get(gomock.Any(), key).Return("", fmt.Errorf("not found"))
	scraperMock.EXPECT().GetLinks(context.Background(), userID).Return(nil, errors.New("database error"))
	mockCtx.EXPECT().Send("Произошла ошибка при получении списка ссылок.").Return(nil)

	domain.Users = map[int64]domain.User{userID: {}}

	err := bot.List(mockCtx)
	assert.Error(t, err)
}
