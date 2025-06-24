package kafkaredis_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/tests/integration/kafka_redis/utils"
)

type TestSuite struct {
	suite.Suite
	kafkaContainer *utils.KafkaContainer
	redisContainer *utils.RedisContainer
}

func (s *TestSuite) SetupSuite() {
	startTime := time.Now()

	log.Println("Starting TestSuite setup")

	defer func() {
		log.Printf("TestSuite setup completed in %v", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var err error
	s.kafkaContainer, err = utils.NewKafkaContainer(ctx)
	s.Require().NoError(err)
	log.Println("Kafka container created")

	s.redisContainer, err = utils.NewRedisContainer(ctx)
	s.Require().NoError(err)
	log.Println("Redis container created")

	topics := []string{"link-updates", "dlq-topic"}
	for _, topic := range topics {
		log.Printf("Creating Kafka topic: %s", topic)
		err = s.createKafkaTopic(ctx, topic, 1)
		s.Require().NoError(err, "Failed to create topic %s", topic)
		log.Printf("Topic %s created successfully", topic)
	}
}

func (s *TestSuite) SetupTest() {
	log.Printf("Preparing test %s...", s.T().Name())
}

func (s *TestSuite) TearDownSuite() {
	log.Println("Starting TestSuite teardown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Terminating containers...")

	if s.kafkaContainer != nil {
		err := s.kafkaContainer.Container.Terminate(ctx)
		s.Require().NoError(err)
		log.Println("Kafka container terminated")

		err = s.kafkaContainer.Zookeeper.Terminate(ctx)
		s.Require().NoError(err)
		log.Println("Zookeeper container terminated")
	}

	if s.redisContainer != nil {
		err := s.redisContainer.Container.Terminate(ctx)
		s.Require().NoError(err)
		log.Println("Redis container terminated")
	}
}

func (s *TestSuite) createKafkaTopic(_ context.Context, topic string, partitions int) error {
	conn, err := kafka.Dial("tcp", s.kafkaContainer.BrokerURL)
	if err != nil {
		return fmt.Errorf("failed to dial Kafka: %w", err)
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	return nil
}

func TestSuite_Run(t *testing.T) {
	log.Println("Starting test suite execution")
	suite.Run(t, new(TestSuite))
	log.Println("Test suite execution completed")
}
