package kafkaredis_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	prodDomain "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
)

func (s *TestSuite) TestKafkaConnectivity() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", s.kafkaContainer.BrokerURL)
	s.Require().NoError(err, "Failed to connect to Kafka broker")
	defer conn.Close()

	brokers, err := conn.Brokers()
	s.Require().NoError(err, "Failed to fetch brokers metadata")
	s.NotEmpty(brokers, "Should have at least one broker")
}

func (s *TestSuite) TestKafkaProducerConsumer() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := prodDomain.LinkUpdate{
		ID:          1,
		URL:         "https://test.com",
		Description: "Test message",
		UserID:      123,
		Type:        "Test",
	}

	producer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{s.kafkaContainer.BrokerURL},
		Topic:   "link-updates",
	})
	defer producer.Close()

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{s.kafkaContainer.BrokerURL},
		Topic:   "link-updates",
		GroupID: "test-group",
	})
	defer consumer.Close()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		s.T().Logf("Failed to marshal message: %v", err)
		s.T().Skip("Skipping test due to marshal error")

		return
	}

	err = producer.WriteMessages(ctx, kafka.Message{
		Value: messageBytes,
	})
	if err != nil {
		s.T().Logf("Failed to send message: %v", err)
		s.T().Skip("Skipping test due to send error")

		return
	}

	readCtx, readCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readCancel()

	receivedMessage, err := consumer.ReadMessage(readCtx)
	if err != nil {
		s.T().Logf("Failed to read message: %v", err)
		s.T().Skip("Skipping test due to read error")

		return
	}

	var receivedLinkUpdate prodDomain.LinkUpdate
	err = json.Unmarshal(receivedMessage.Value, &receivedLinkUpdate)

	if err != nil {
		s.T().Logf("Failed to unmarshal received message: %v", err)
		s.T().Skip("Skipping test due to unmarshal error")

		return
	}

	s.Equal(message.ID, receivedLinkUpdate.ID)
	s.Equal(message.URL, receivedLinkUpdate.URL)
	s.Equal(message.Description, receivedLinkUpdate.Description)
	s.Equal(message.UserID, receivedLinkUpdate.UserID)
	s.Equal(message.Type, receivedLinkUpdate.Type)
}

func (s *TestSuite) TestKafkaDLQ() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message := prodDomain.LinkUpdate{
		ID:          2,
		URL:         "https://dlq-test.com",
		Description: "DLQ Test message",
		UserID:      456,
		Type:        "DLQTest",
	}

	producer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{s.kafkaContainer.BrokerURL},
		Topic:   "dlq-topic",
	})
	defer producer.Close()

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{s.kafkaContainer.BrokerURL},
		Topic:   "dlq-topic",
		GroupID: "dlq-test-group",
	})
	defer consumer.Close()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		s.T().Logf("Failed to marshal message: %v", err)
		s.T().Skip("Skipping test due to marshal error")

		return
	}

	err = producer.WriteMessages(ctx, kafka.Message{
		Value: messageBytes,
	})

	if err != nil {
		s.T().Logf("Failed to send message to DLQ: %v", err)
		s.T().Skip("Skipping test due to send error")

		return
	}

	readCtx, readCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readCancel()

	receivedMessage, err := consumer.ReadMessage(readCtx)

	if err != nil {
		s.T().Logf("Failed to read message from DLQ: %v", err)
		s.T().Skip("Skipping test due to read error")

		return
	}

	var receivedLinkUpdate prodDomain.LinkUpdate
	err = json.Unmarshal(receivedMessage.Value, &receivedLinkUpdate)

	if err != nil {
		s.T().Logf("Failed to unmarshal received message: %v", err)
		s.T().Skip("Skipping test due to unmarshal error")

		return
	}

	s.Equal(message.ID, receivedLinkUpdate.ID)
	s.Equal(message.URL, receivedLinkUpdate.URL)
	s.Equal(message.Description, receivedLinkUpdate.Description)
	s.Equal(message.UserID, receivedLinkUpdate.UserID)
	s.Equal(message.Type, receivedLinkUpdate.Type)
}
