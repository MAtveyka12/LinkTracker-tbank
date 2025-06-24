package telegram

import "gopkg.in/telebot.v3"

func (t *TgBot) HandleUnknownCommand(ctx telebot.Context) error {
	return ctx.Send("Неизвестная команда. Введите /help для списка доступных команд.")
}
