package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	testcontainers.Container
	Config PostgreSQLContainerConfig
}

type PostgreSQLContainerOption func(c *PostgreSQLContainerConfig)

type PostgreSQLContainerConfig struct {
	ImageTag   string
	User       string
	Password   string
	MappedPort string
	Database   string
	Host       string
}

func (c *PostgreSQLContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Config.User, c.Config.Password, c.Config.Host, c.Config.MappedPort, c.Config.Database)
}

func NewPostgreSQLContainer(ctx context.Context, opts ...PostgreSQLContainerOption) (*PostgreSQLContainer, error) {
	log.Println("Starting PostgreSQL container initialization")

	const (
		psqlImage = "postgres"
		psqlPort  = "5432"
	)

	config := PostgreSQLContainerConfig{
		ImageTag: "latest",
		User:     "postgres",
		Password: "postgres",
		Database: "track_link_tbank",
	}
	log.Printf("Default container config: %+v", config)

	for _, opt := range opts {
		opt(&config)
	}

	log.Printf("Final container config after options: %+v", config)

	containerPort := psqlPort + "/tcp"
	log.Printf("Using container port: %s", containerPort)

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"POSTGRES_USER":     config.User,
				"POSTGRES_PASSWORD": config.Password,
				"POSTGRES_DB":       config.Database,
			},
			ExposedPorts: []string{
				containerPort,
			},
			Image:      fmt.Sprintf("%s:%s", psqlImage, config.ImageTag),
			WaitingFor: wait.ForListeningPort(nat.Port(containerPort)).WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		return nil, fmt.Errorf("creating container: %w", err)
	}

	log.Println("Container created successfully")

	host, err := container.Host(ctx)
	if err != nil {
		log.Printf("Failed to get container host: %v", err)
		return nil, fmt.Errorf("getting host: %w", err)
	}

	log.Printf("Container host: %s", host)

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		log.Printf("Failed to get mapped port: %v", err)
		return nil, fmt.Errorf("getting mapped port: %w", err)
	}

	log.Printf("Mapped port: %s", mappedPort.Port())

	config.MappedPort = mappedPort.Port()
	config.Host = host

	log.Println("PostgreSQL container initialized successfully")

	return &PostgreSQLContainer{Container: container, Config: config}, nil
}
