package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	Container testcontainers.Container
	Addr      string
}

func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	log.Println("Starting Redis container initialization")

	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7.2",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Redis: %w", err)
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis host: %w", err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis port: %w", err)
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort.Port())
	log.Printf("Redis server running at: %s", addr)

	return &RedisContainer{
		Container: redisContainer,
		Addr:      addr,
	}, nil
}
