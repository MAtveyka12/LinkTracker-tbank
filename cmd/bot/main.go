package main

import (
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/app"
)

func main() {
	appBot, err := app.New()

	if err != nil {
		slog.Error(err.Error())
		return
	}

	if err := appBot.Run(); err != nil {
		slog.Error(err.Error())
		return
	}
}
