package telegram

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
	"gopkg.in/telebot.v3"
)

type LinkUpdate struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	UserID      int64  `json:"user_id"`
	Type        string `json:"type"`
}

func (t *TgBot) KafkaMessageHandler(msg *sarama.ConsumerMessage) error {
	var update LinkUpdate

	if err := json.Unmarshal(msg.Value, &update); err != nil {
		t.logger.Error("Failed to unmarshal Kafka message", slog.String("error", err.Error()))
		return err
	}

	message := formatUpdateMessageKafka(update)

	_, err := t.Bot.Send(telebot.ChatID(update.UserID), message)
	if err != nil {
		t.logger.Error("Failed to send Telegram message", slog.String("error", err.Error()))
		return err
	}

	t.logger.Info("Message sent to Telegram successfully")

	return nil
}

func formatUpdateMessageKafka(update LinkUpdate) string {
	return "Новое обновление!\n" +
		update.URL + "\n" +
		update.Description
}
