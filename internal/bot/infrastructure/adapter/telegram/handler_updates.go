package telegram

import (
	"encoding/json"
	"net/http"

	"gopkg.in/telebot.v3"
)

func UpdatesHandler(bot *TgBot) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update LinkUpdate

		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		message := formatUpdateMessage(update)

		_, err := bot.Bot.Send(telebot.ChatID(update.UserID), message)

		if err != nil {
			http.Error(w, "error sending the message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func formatUpdateMessage(update LinkUpdate) string {
	return "Новое обновление!\n" +
		update.URL + "\n" +
		update.Description
}
