package kafka

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/IBM/sarama"
)

type Consumer struct {
	topic      string
	group      string
	updateChan chan sarama.ConsumerMessage
	stopCh     chan struct{}
	logger     *slog.Logger
	client     sarama.ConsumerGroup
	wg         sync.WaitGroup
}

func NewKafkaConsumer(
	addresses []string,
	topic, group string,
	updateCh chan sarama.ConsumerMessage,
	logger *slog.Logger,
	config *sarama.Config,
) (*Consumer, error) {
	const op = "KafkaConsumer.New"
	logger = logger.With(
		slog.String("op", op),
		slog.String("topic", topic),
		slog.String("group", group),
		slog.Any("brokers", addresses),
	)

	client, err := sarama.NewConsumerGroup(addresses, group, config)
	if err != nil {
		logger.Error("Failed to create consumer group", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("kafka consumer created successfully",
		slog.Int("brokers_count", len(addresses)))

	return &Consumer{
		topic:      topic,
		group:      group,
		updateChan: updateCh,
		stopCh:     make(chan struct{}),
		logger:     logger,
		client:     client,
	}, nil
}

func (kc *Consumer) Run(ctx context.Context) error {
	const op = "KafkaConsumer.Run"
	logger := kc.logger.With(slog.String("op", op))

	kc.wg.Add(1)

	consumeMessages := func() {
		defer kc.wg.Done()

		defer func() {
			if err := kc.client.Close(); err != nil {
				logger.Error("failed to close consumer group",
					slog.String("error", err.Error()))
			}

			logger.Info("consumer group closed")
		}()

		for {
			kc.logger.Info("Calling client.Consume()...")

			if err := kc.client.Consume(ctx, []string{kc.topic}, kc); err != nil {
				kc.logger.Error("Error during consuming", slog.String("error", err.Error()))
			}

			if ctx.Err() != nil {
				return
			}
		}
	}

	go consumeMessages()

	kc.wg.Add(1)

	shutdownHandler := func() {
		defer kc.wg.Done()
		kc.handleShutdown(ctx)
	}
	go shutdownHandler()

	kc.logger.Info("Kafka consumer group started", slog.String("topic", kc.topic), slog.String("group", kc.group))

	return nil
}

func (kc *Consumer) handleShutdown(ctx context.Context) {
	const op = "KafkaConsumer.handleShutdown"
	logger := kc.logger.With(slog.String("op", op))

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

	select {
	case <-kc.stopCh:
		logger.Info("received internal stop signal")
	case <-sigterm:
		logger.Info("received OS interrupt signal")
	case <-ctx.Done():
		logger.Info("context canceled")
	}

	logger.Info("initiating shutdown sequence")

	if err := kc.client.Close(); err != nil {
		kc.logger.Error("Error closing consumer group", slog.String("error", err.Error()))
	}
}

func (kc *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	kc.logger.Info("Consumer group setup complete")
	return nil
}

func (kc *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	kc.logger.Info("Consumer group cleanup complete")
	return nil
}

func (kc *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	kc.logger.Info("ConsumeClaim called!")

	const op = "KafkaConsumer.ConsumeClaim"
	logger := kc.logger.With(
		slog.String("op", op),
		slog.String("topic", claim.Topic()),
		slog.String("member_id", session.MemberID()),
	)

	logger.Info("starting to consume claims",
		slog.Int64("initial_offset", claim.InitialOffset()))

	for message := range claim.Messages() {
		kc.logger.Info("Message received", slog.String("topic", message.Topic), slog.Int64("offset", message.Offset))
		kc.updateChan <- *message
		session.MarkMessage(message, "")
	}

	return nil
}

func (kc *Consumer) Stop() {
	const op = "KafkaConsumer.Stop"

	kc.logger.With(slog.String("op", op)).Info("initiating consumer stop")
	close(kc.stopCh)
}

func (kc *Consumer) Messages() <-chan sarama.ConsumerMessage {
	return kc.updateChan
}

func (kc *Consumer) Wait() {
	kc.wg.Wait()
}
