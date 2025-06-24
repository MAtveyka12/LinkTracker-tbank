package kafka

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/IBM/sarama"

	prodDomain "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer"
)

type Producer struct {
	logger     *slog.Logger
	producer   sarama.AsyncProducer
	chanUpdate chan prodDomain.LinkUpdate
	topic      string
	dlqTopic   string
	stopCh     chan struct{}
	doneCh     chan struct{}
}

func NewKafkaProducer(
	addresses []string,
	logger *slog.Logger,
	update chan prodDomain.LinkUpdate,
	topic string,
	dlqTopic string,
	saramaConfig *sarama.Config,
) (producer.ProdKafkaOrHTTP, error) {
	const op = "KafkaProducer.New"
	logger = logger.With(slog.String("op", op), slog.String("topic", topic))

	prod, err := sarama.NewAsyncProducer(addresses, saramaConfig)
	if err != nil {
		logger.Error("failed to create a new async Kafka producer",
			slog.String("error", err.Error()),
			slog.Any("brokers", addresses))

		return nil, err
	}

	kp := &Producer{
		logger:     logger,
		producer:   prod,
		chanUpdate: update,
		topic:      topic,
		dlqTopic:   dlqTopic,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}

	go kp.handleDeliveryResults()

	logger.Info("kafka producer created successfully",
		slog.Int("brokers_count", len(addresses)))

	return kp, nil
}

func (kp *Producer) Run() {
	const op = "KafkaProducer.Run"

	kp.logger.Info(op, slog.String("msg", "Kafka producer is running"))

	go func() {
		defer func() {
			if err := kp.producer.Close(); err != nil {
				kp.logger.Error("failed to close producer", slog.String("error", err.Error()))
			}

			kp.logger.Info("kafka producer stopped")
		}()

		for {
			select {
			case update, ok := <-kp.chanUpdate:
				if !ok {
					kp.logger.Warn(op, slog.String("msg", "Update channel closed, stopping producer"))
					return
				}

				topicUserID := strconv.Itoa(int(update.UserID))

				if err := kp.produceUpdate(update); err != nil {
					kp.logger.Error(op, slog.String("topic", topicUserID), slog.String("error", err.Error()))
				} else {
					kp.logger.Info(op,
						slog.String("user_id", topicUserID),
						slog.String("status", "Produced successfully"),
						slog.String("message_type", update.Type))
				}

			case <-kp.stopCh:
				kp.logger.Warn(op, slog.String("msg", "Stopping Kafka producer"))
				return
			}
		}
	}()
}

func (kp *Producer) Stop() {
	const op = "KafkaProducer.Stop"

	kp.logger.Warn(op, slog.String("msg", "Stopping Kafka producer"))

	kp.logger.Info("initiating graceful shutdown")

	kp.producer.AsyncClose()
	close(kp.stopCh)
	<-kp.doneCh
}

func (kp *Producer) produceUpdate(update prodDomain.LinkUpdate) error {
	const op = "KafkaProducer.produceUpdate"

	kp.logger.Info("Sending message to Kafka",
		slog.Int64("id", update.ID),
		slog.String("url", update.URL),
		slog.String("description", update.Description),
		slog.Int64("user_id", update.UserID),
		slog.String("type", update.Type),
	)

	commitJSON, err := json.Marshal(update)
	if err != nil {
		kp.logger.Error(op, slog.String("msg", "Failed to marshal"), slog.String("error", err.Error()))

		return kp.sendToDLQ(update, "marshal_failed", err.Error())
	}

	msg := &sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d-%s", update.UserID, update.Type)),
		Value: sarama.ByteEncoder(commitJSON),
	}

	kp.producer.Input() <- msg

	return nil
}

func (kp *Producer) sendToDLQ(update prodDomain.LinkUpdate, reason, errMsg string) error {
	const op = "KafkaProducer.sendToDLQ"

	dlqContent := map[string]interface{}{
		"reason":  reason,
		"error":   errMsg,
		"payload": update,
	}

	payload, err := json.Marshal(dlqContent)
	if err != nil {
		kp.logger.Error(op, slog.String("msg", "Failed to marshal DLQ payload"), slog.String("error", err.Error()))
		return err
	}

	dlqMsg := &sarama.ProducerMessage{
		Topic: kp.dlqTopic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d-%s", update.UserID, update.Type)),
		Value: sarama.ByteEncoder(payload),
	}

	kp.producer.Input() <- dlqMsg

	kp.logger.Warn(op, slog.String("msg", "Sent message to DLQ"), slog.Any("payload", update))

	return nil
}

func (kp *Producer) handleDeliveryResults() {
	for {
		select {
		case success := <-kp.producer.Successes():
			kp.logger.Info("Message delivered",
				slog.String("topic", success.Topic),
				slog.Int64("partition", int64(success.Partition)),
				slog.Int64("offset", success.Offset))
		case err := <-kp.producer.Errors():
			kp.logger.Error("Failed to deliver message",
				slog.String("error", err.Err.Error()),
				slog.Any("msg", err.Msg))

			go kp.retryMessage(err.Msg, 3)

		case <-kp.stopCh:
			close(kp.doneCh)
			return
		}
	}
}

func (kp *Producer) retryMessage(msg *sarama.ProducerMessage, maxRetries int) {
	const op = "KafkaProducer.retryMessage"

	for i := 0; i < maxRetries; i++ {
		time.Sleep(time.Second * time.Duration(i+1))

		kp.logger.Info(op, slog.Int("attempt", i+1), slog.String("status", "Retrying message"))

		select {
		case kp.producer.Input() <- msg:
			kp.logger.Info(op, slog.String("msg", "Retry succeeded"))

			return
		default:
			kp.logger.Warn(op, slog.String("msg", "Retry attempt failed"))
		}
	}

	kp.logger.Error(op, slog.String("msg", "Max retries reached. Message dropped"))
}

func (kp *Producer) SendUpdate(_ prodDomain.LinkUpdate) error { return nil }

func (kp *Producer) SendUpdatePR(_ prodDomain.LinkUpdate) error { return nil }
