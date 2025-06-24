package telegram

import (
	"fmt"
	"os"

	"gopkg.in/telebot.v3"
)

func (t *TgBot) Help(ctx telebot.Context) error {
	templatePath := os.Getenv("TMPL_HELP_FILE_PATH")
	if templatePath == "" {
		return fmt.Errorf("TMPL_HELP_FILE_PATH environment variable not set")
	}

	content, err := os.ReadFile(templatePath)

	if err != nil {
		return ctx.Send(fmt.Errorf("ошибка загрузки списка команд: %s", err.Error()))
	}

	return ctx.Send(string(content))
}

func (t *TgBot) HelpHandler() {
	t.Bot.Handle("/help", t.Help)
}
