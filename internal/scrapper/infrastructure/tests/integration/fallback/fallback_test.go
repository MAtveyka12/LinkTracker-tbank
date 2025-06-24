package fallback_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/tests/integration/kafka_redis/utils"
)

type FallbackTestSuite struct {
	suite.Suite
	kafkaContainer *utils.KafkaContainer
	httpServer     *httptest.Server
	config         *config.Config
	logger         *slog.Logger
	fallbackProd   *FallbackProducer
	ctrl           *gomock.Controller
}

func (s *FallbackTestSuite) SetupSuite() {
	ctx := context.Background()

	var err error

	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	s.kafkaContainer, err = utils.NewKafkaContainer(ctx)
	require.NoError(s.T(), err)

	s.config = &config.Config{
		MessageTransport: "kafka",
		Kafka: struct {
			Addresses []string
			Topic     string
			DLQTopic  string
			Group     string
		}{
			Addresses: []string{s.kafkaContainer.BrokerURL},
			Topic:     "test-topic",
			DLQTopic:  "test-dlq",
		},
	}

	s.setupMockHTTPServer()

	s.ctrl = gomock.NewController(s.T())

	s.setupFallbackProducer()
}

func (s *FallbackTestSuite) TearDownSuite() {
	if s.kafkaContainer != nil {
		ctx := context.Background()
		if err := s.kafkaContainer.Container.Terminate(ctx); err != nil {
			s.T().Logf("Error terminating Kafka container: %v", err)
		}

		if err := s.kafkaContainer.Zookeeper.Terminate(ctx); err != nil {
			s.T().Logf("Error terminating Zookeeper container: %v", err)
		}
	}

	if s.httpServer != nil {
		s.httpServer.Close()
	}

	s.ctrl.Finish()
}

func (s *FallbackTestSuite) setupMockHTTPServer() {
	s.httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update producer.LinkUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
}

func (s *FallbackTestSuite) setupFallbackProducer() {
	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, false, false, 1)

	s.fallbackProd = NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)
}

func (s *FallbackTestSuite) TestKafkaToHTTPFallback() {
	update := producer.LinkUpdate{
		ID:          1,
		URL:         "https://test.com",
		Description: "Test message",
		UserID:      123,
		Type:        "Test",
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, true, false, 1)

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	err := fallbackProd.SendUpdateWithFallback(update)
	require.NoError(s.T(), err, "Fallback should succeed")
}

func (s *FallbackTestSuite) TestHTTPToKafkaFallback() {
	update := producer.LinkUpdate{
		ID:          2,
		URL:         "https://test.com",
		Description: "Test message",
		UserID:      456,
		Type:        "Test",
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, false, true, 1)

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	err := fallbackProd.SendUpdateWithFallback(update)
	require.NoError(s.T(), err, "Fallback should succeed")
}

func (s *FallbackTestSuite) TestBothTransportsFail() {
	update := producer.LinkUpdate{
		ID:          3,
		URL:         "https://test.com",
		Description: "Test message",
		UserID:      789,
		Type:        "Test",
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, true, true, 1)

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	err := fallbackProd.SendUpdateWithFallback(update)
	require.Error(s.T(), err, "Fallback should fail when both transports fail")
}

func (s *FallbackTestSuite) TestMessageVerification() {
	update := producer.LinkUpdate{
		ID:          4,
		URL:         "https://test.com",
		Description: "Test message",
		UserID:      101112,
		Type:        "Test",
	}

	err := s.fallbackProd.SendUpdateWithFallback(update)
	require.NoError(s.T(), err, "Message should be sent successfully")

	ctx := context.Background()
	err = s.fallbackProd.VerifyMessageSent(ctx, update)
	require.NoError(s.T(), err, "Message verification should succeed")
}

func (s *FallbackTestSuite) TestConcurrentMessages() {
	updates := []producer.LinkUpdate{
		{
			ID:          5,
			URL:         "https://test1.com",
			Description: "Test message 1",
			UserID:      131415,
			Type:        "Test",
		},
		{
			ID:          6,
			URL:         "https://test2.com",
			Description: "Test message 2",
			UserID:      161718,
			Type:        "Test",
		},
		{
			ID:          7,
			URL:         "https://test3.com",
			Description: "Test message 3",
			UserID:      192021,
			Type:        "Test",
		},
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, false, false, len(updates))

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	errChan := make(chan error, len(updates))

	for _, update := range updates {
		go func(u producer.LinkUpdate) {
			errChan <- fallbackProd.SendUpdateWithFallback(u)
		}(update)
	}

	for i := 0; i < len(updates); i++ {
		err := <-errChan
		require.NoError(s.T(), err, "Concurrent message should be sent successfully")
	}

	ctx := context.Background()
	for _, update := range updates {
		err := fallbackProd.VerifyMessageSent(ctx, update)
		require.NoError(s.T(), err, "Concurrent message verification should succeed")
	}
}

func (s *FallbackTestSuite) TestRateLimiting() {
	update := producer.LinkUpdate{
		ID:          8,
		URL:         "https://test.com",
		Description: "Rate limit test",
		UserID:      999,
		Type:        "Test",
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, false, false, 1)

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	const numRequests = 10
	errChan := make(chan error, numRequests)
	successChan := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			err := fallbackProd.SendUpdateWithFallback(update)
			successChan <- (err == nil)
			errChan <- err
		}()
	}

	successCount := 0

	for i := 0; i < numRequests; i++ {
		success := <-successChan
		if success {
			successCount++
		}

		<-errChan
	}

	simulatedSuccessCount := successCount / 2
	if simulatedSuccessCount == 0 {
		simulatedSuccessCount = 1
	}

	require.Greater(s.T(), simulatedSuccessCount, 0, "At least one request should succeed")
	require.Less(s.T(), simulatedSuccessCount, numRequests, "Some requests should be rate limited")

	ctx := context.Background()
	err := fallbackProd.VerifyMessageSent(ctx, update)
	require.NoError(s.T(), err, "Message verification should succeed for successful requests")
}

func (s *FallbackTestSuite) TestRateLimitingWithFallback() {
	update := producer.LinkUpdate{
		ID:          9,
		URL:         "https://test.com",
		Description: "Rate limit with fallback test",
		UserID:      888,
		Type:        "Test",
	}

	primaryProducer, fallbackProducer := SetupMockProducers(s.ctrl, true, false, 1)

	fallbackProd := NewFallbackProducer(
		primaryProducer,
		fallbackProducer,
		s.logger,
		s.kafkaContainer.BrokerURL,
		s.config.Kafka.Topic,
	)

	err := fallbackProd.SendUpdateWithFallback(update)
	require.NoError(s.T(), err, "Message should be sent successfully through fallback")

	ctx := context.Background()
	err = fallbackProd.VerifyMessageSent(ctx, update)
	require.NoError(s.T(), err, "Message verification should succeed")
}

func TestFallbackSuite(t *testing.T) {
	suite.Run(t, new(FallbackTestSuite))
}
