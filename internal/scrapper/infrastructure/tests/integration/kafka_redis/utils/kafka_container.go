package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type KafkaContainer struct {
	Container testcontainers.Container
	Zookeeper testcontainers.Container
	BrokerURL string
}

func NewKafkaContainer(ctx context.Context) (*KafkaContainer, error) {
	log.Println("Starting Kafka container initialization")

	net, err := createNetwork(ctx)
	if err != nil {
		return nil, err
	}

	zookeeper, err := startZookeeper(ctx, net.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to start Zookeeper: %w", err)
	}

	kafka, brokerURL, err := startKafka(ctx, net.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to start Kafka: %w", err)
	}

	return &KafkaContainer{
		Container: kafka,
		Zookeeper: zookeeper,
		BrokerURL: brokerURL,
	}, nil
}

func createNetwork(ctx context.Context) (*testcontainers.DockerNetwork, error) {
	net, err := network.New(ctx,
		network.WithDriver("bridge"),
		network.WithAttachable(),
		network.WithLabels(map[string]string{
			"test": "kafka-network",
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	return net, nil
}

func startZookeeper(ctx context.Context, networkName string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-zookeeper:7.3.0",
		ExposedPorts: []string{"2181/tcp"},
		WaitingFor:   wait.ForListeningPort("2181/tcp").WithStartupTimeout(2 * time.Minute),
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
			"ZOOKEEPER_TICK_TIME":   "2000",
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"zookeeper"},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Zookeeper: %s", err.Error())
	}

	return container, nil
}

func startKafka(ctx context.Context, networkName string) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.3.0",
		ExposedPorts: []string{"9093/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9093/tcp").WithStartupTimeout(2*time.Minute),
			wait.ForLog("started (kafka.server.KafkaServer)"),
		).WithDeadline(2 * time.Minute),
		Env: map[string]string{
			"KAFKA_ZOOKEEPER_CONNECT":                        "zookeeper:2181",
			"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://kafka:9092,EXTERNAL://localhost:9093",
			"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:9092,EXTERNAL://0.0.0.0:9093",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,EXTERNAL:PLAINTEXT",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "PLAINTEXT",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
			"KAFKA_BROKER_ID":                                "1",
			"KAFKA_LOG_DIRS":                                 "/tmp/kafka-logs",
			"KAFKA_LOG_RETENTION_HOURS":                      "1",
			"KAFKA_LOG_RETENTION_CHECK_INTERVAL_MS":          "300000",
			"KAFKA_LOG_SEGMENT_BYTES":                        "1073741824",
			"KAFKA_LOG_CLEANUP_POLICY":                       "delete",
			"KAFKA_NUM_PARTITIONS":                           "1",
			"KAFKA_DEFAULT_REPLICATION_FACTOR":               "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_OFFSETS_TOPIC_NUM_PARTITIONS":             "1",
			"KAFKA_TRANSACTION_MAX_TIMEOUT_MS":               "900000",
			"KAFKA_MAX_POLL_INTERVAL_MS":                     "300000",
			"KAFKA_SESSION_TIMEOUT_MS":                       "10000",
			"KAFKA_HEARTBEAT_INTERVAL_MS":                    "3000",
			"KAFKA_DELETE_TOPIC_ENABLE":                      "true",
			"KAFKA_OFFSETS_RETENTION_MINUTES":                "1",
			"KAFKA_OFFSETS_RETENTION_CHECK_INTERVAL_MS":      "300000",
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"kafka"},
		},
		SkipReaper: false,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start Kafka: %s", err.Error())
	}

	port, err := container.MappedPort(ctx, "9093/tcp")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get Kafka port: %s", err.Error())
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get Kafka host: %s", err.Error())
	}

	brokerURL := fmt.Sprintf("%s:%s", host, port.Port())
	log.Printf("Kafka broker running at: %s", brokerURL)

	return container, brokerURL, nil
}
