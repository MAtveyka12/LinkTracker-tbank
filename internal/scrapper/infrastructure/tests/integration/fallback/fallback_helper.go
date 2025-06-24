package fallback_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/mock/gomock"

	prodDomain "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	mock "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer/mock"
)

type FallbackProducer struct {
	primaryProducer  *mock.MockProdKafkaOrHTTP
	fallbackProducer *mock.MockProdKafkaOrHTTP
	logger           *slog.Logger
	brokerURL        string
	topic            string
}

func NewFallbackProducer(
	primaryProducer *mock.MockProdKafkaOrHTTP,
	fallbackProducer *mock.MockProdKafkaOrHTTP,
	logger *slog.Logger,
	brokerURL string,
	topic string,
) *FallbackProducer {
	return &FallbackProducer{
		primaryProducer:  primaryProducer,
		fallbackProducer: fallbackProducer,
		logger:           logger,
		brokerURL:        brokerURL,
		topic:            topic,
	}
}

func (fp *FallbackProducer) SendUpdateWithFallback(update prodDomain.LinkUpdate) error {
	err := fp.primaryProducer.SendUpdate(update)
	if err != nil {
		fp.logger.Warn("Primary transport failed, falling back to secondary",
			slog.String("error", err.Error()),
			slog.Any("update", update),
		)

		return fp.fallbackProducer.SendUpdate(update)
	}

	return nil
}

func (fp *FallbackProducer) VerifyMessageSent(ctx context.Context, update prodDomain.LinkUpdate) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{fp.brokerURL},
		Topic:   fp.topic,
		GroupID: "test-verification-group",
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		return nil
	}

	var receivedUpdate prodDomain.LinkUpdate
	if err := json.Unmarshal(msg.Value, &receivedUpdate); err != nil {
		return fmt.Errorf("failed to unmarshal received message: %w", err)
	}

	if receivedUpdate.ID != update.ID ||
		receivedUpdate.URL != update.URL ||
		receivedUpdate.Description != update.Description ||
		receivedUpdate.UserID != update.UserID ||
		receivedUpdate.Type != update.Type {
		return fmt.Errorf("received message does not match sent message")
	}

	return nil
}

func SetupMockProducers(
	ctrl *gomock.Controller,
	primaryShouldFail bool,
	fallbackShouldFail bool,
	_ int,
) (primaryProducer, fallbackProducer *mock.MockProdKafkaOrHTTP) {
	primaryProducer = mock.NewMockProdKafkaOrHTTP(ctrl)
	fallbackProducer = mock.NewMockProdKafkaOrHTTP(ctrl)

	if primaryShouldFail {
		primaryProducer.EXPECT().
			SendUpdate(gomock.Any()).
			Return(fmt.Errorf("mock primary producer failure")).
			AnyTimes()
	} else {
		primaryProducer.EXPECT().
			SendUpdate(gomock.Any()).
			Return(nil).
			AnyTimes()
	}

	if fallbackShouldFail {
		fallbackProducer.EXPECT().
			SendUpdate(gomock.Any()).
			Return(fmt.Errorf("mock fallback producer failure")).
			AnyTimes()
	} else {
		fallbackProducer.EXPECT().
			SendUpdate(gomock.Any()).
			Return(nil).
			AnyTimes()
	}

	return primaryProducer, fallbackProducer
}
