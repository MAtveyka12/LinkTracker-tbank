package telegram

import (
	"log/slog"

	"gopkg.in/telebot.v3"
)

func (t *TgBot) RegisterCommands() {
	commands := []telebot.Command{
		{Text: "start", Description: "Регистрация пользователя"},
		{Text: "help", Description: "Вывод списка доступных команд"},
		{Text: "track", Description: "Начать отслеживание ссылки"},
		{Text: "untrack", Description: "Прекратить отслеживание ссылки"},
		{Text: "list", Description: "Показать список отслеживаемых ссылок"},
	}

	if err := t.Bot.SetCommands(commands); err != nil {
		slog.Error("Error when registering commands: %s", "error", err.Error())
	}
}
