package application

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/adapter/telegram"
)

func SetupRouter(t *telegram.TgBot) {
	t.StartHandler()
	t.MessageHandler()
	t.HelpHandler()
	t.TrackHandler()
	t.UntrackHandler()
	t.ListHandler()
	t.RegisterCommands()
}
