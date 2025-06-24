package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	prod "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer"
)

type Producer struct {
	botAPIURL string
	client    *http.Client
}

func NewHTTPProducer() producer.ProdKafkaOrHTTP {
	return &Producer{
		botAPIURL: "http://bot:7070",
		client:    &http.Client{},
	}
}

func (p *Producer) SendUpdate(update prod.LinkUpdate) error {
	slog.Debug("Sending link update",
		slog.String("operation", "send_update"),
		slog.Any("update", map[string]interface{}{
			"chat_id": update.ID,
			"url":     update.URL,
		}),
	)

	jsonData, err := json.Marshal(update)
	if err != nil {
		slog.Error("Failed to marshal update",
			slog.String("error", err.Error()),
			slog.Any("update", update),
		)

		return fmt.Errorf("data serialization error: %s", err.Error())
	}

	req, err := http.NewRequest("POST", p.botAPIURL+"/updates", bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create request",
			slog.String("error", err.Error()),
			slog.String("url", p.botAPIURL+"/updates"),
		)

		return fmt.Errorf("request creation error: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	slog.Debug("Sending HTTP request",
		slog.String("url", p.botAPIURL+"/updates"),
		slog.Int("data_length", len(jsonData)),
	)

	resp, err := p.client.Do(req)
	if err != nil {
		slog.Error("HTTP request failed",
			slog.String("error", err.Error()),
			slog.String("url", p.botAPIURL+"/updates"),
		)

		return fmt.Errorf("request execution error: %s", err.Error())
	}

	defer resp.Body.Close()

	slog.Debug("Received HTTP response",
		slog.Int("status_code", resp.StatusCode),
		slog.String("status", resp.Status),
	)

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code",
			slog.Int("status_code", resp.StatusCode),
			slog.String("expected_code", "200"),
		)

		return fmt.Errorf("error sending notification: %d", resp.StatusCode)
	}

	slog.Info("Update sent successfully",
		slog.Int("chat_id", int(update.ID)),
		slog.String("url", update.URL),
	)

	return nil
}

func (p *Producer) SendUpdatePR(update prod.LinkUpdate) error {
	slog.Debug("Sending PR update",
		slog.String("operation", "send_pr_update"),
		slog.Any("update", map[string]interface{}{
			"chat_id":       update.ID,
			"pr_url":        update.URL,
			"repo_owner":    update.UserID,
			"repo_descript": update.Description,
		}),
	)

	jsonData, err := json.Marshal(update)
	if err != nil {
		slog.Error("Failed to marshal PR update",
			slog.String("error", err.Error()),
			slog.Any("update", update),
		)

		return fmt.Errorf("data serialization error: %s", err.Error())
	}

	req, err := http.NewRequest("POST", p.botAPIURL+"/updates", bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create request",
			slog.String("error", err.Error()),
			slog.String("url", p.botAPIURL+"/updates"),
		)

		return fmt.Errorf("request creation error: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	slog.Debug("Sending HTTP request",
		slog.String("url", p.botAPIURL+"/updates"),
		slog.Int("data_length", len(jsonData)),
	)

	resp, err := p.client.Do(req)
	if err != nil {
		slog.Error("HTTP request failed",
			slog.String("error", err.Error()),
			slog.String("url", p.botAPIURL+"/updates"),
		)

		return fmt.Errorf("request execution error: %s", err.Error())
	}

	defer resp.Body.Close()

	slog.Debug("Received HTTP response",
		slog.Int("status_code", resp.StatusCode),
		slog.String("status", resp.Status),
	)

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code",
			slog.Int("status_code", resp.StatusCode),
			slog.String("expected_code", "200"),
		)

		return fmt.Errorf("error sending notification: %d", resp.StatusCode)
	}

	slog.Info("PR update sent successfully",
		slog.Int("chat_id", int(update.ID)),
		slog.String("pr_url", update.URL),
	)

	return nil
}

func (p *Producer) Run() {}

func (p *Producer) Stop() {}
