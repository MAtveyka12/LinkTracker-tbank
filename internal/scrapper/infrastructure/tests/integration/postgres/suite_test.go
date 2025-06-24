package postgres_test

import (
	"context"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/suite"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/pgxpool"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/tests/integration/postgres/utils"
)

type TestSuite struct {
	suite.Suite
	psqlContainer *utils.PostgreSQLContainer
	pgPool        *pgx.Pool

	txManager repository.TxManager
	linkRepo  repository.LinkRepository
	chatRepo  repository.ChatRepository
}

func (s *TestSuite) SetupSuite() {
	startTime := time.Now()

	log.Println("Starting TestSuite setup")

	defer func() {
		log.Printf("TestSuite setup completed in %v", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	logger := slog.New(tint.NewHandler(os.Stdout, nil))

	log.Println("Logger initialized")

	cfg, err := config.LoadConfig("../../../../../../config/config.yaml", "../../../../../../.env")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	}

	s.Require().NoError(err)

	log.Println("Configuration loaded successfully")
	log.Println("Creating PostgreSQL container...")

	s.psqlContainer, err = utils.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	log.Printf("PostgreSQL container created: %s", s.psqlContainer.GetDSN())
	log.Println("Running migrations...")

	err = utils.RunMigrations(s.psqlContainer.GetDSN(), "../../../../../../migrations/goose")
	s.Require().NoError(err)

	log.Println("Migrations completed")
	log.Println("Creating connection pool...")

	poolConfig, err := pgx.ParseConfig(s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	log.Println("Connection config parsed")

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.HealthCheckPeriod = time.Duration(cfg.Database.HealthCheckPeriod)
	poolConfig.MaxConnIdleTime = time.Duration(cfg.Database.MaxIdleTime)
	poolConfig.MaxConnLifetime = time.Duration(cfg.Database.MaxLifeTime)

	log.Printf("Pool configured: MaxConns=%d, MinConns=%d", poolConfig.MaxConns, poolConfig.MinConns)

	s.pgPool, err = pgx.NewWithConfig(ctx, poolConfig)
	s.Require().NoError(err)

	log.Println("Connection pool created")

	err = s.pgPool.Ping(ctx)
	s.Require().NoError(err, "Failed to ping database")

	ctxManager := pgxpool.NewCtxManager(s.pgPool)

	log.Println("Context manager created")

	s.txManager = pgxpool.NewTxManager(s.pgPool, logger, ctxManager)
	s.linkRepo = pgxpool.NewLinkRepository(ctxManager)
	s.chatRepo = pgxpool.NewChatRepository(ctxManager)

	log.Println("Repositories initialized")
}

func (s *TestSuite) SetupTest() {
	log.Printf("Preparing test %s...", s.T().Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	log.Println("Truncating tables...")

	_, err := s.pgPool.Exec(ctx, "TRUNCATE TABLE links, chats RESTART IDENTITY CASCADE")

	cancel()
	s.Require().NoError(err)

	log.Println("Tables truncated")
}

func (s *TestSuite) TearDownSuite() {
	log.Println("Starting TestSuite teardown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	log.Println("Closing connection pool...")

	s.pgPool.Close()

	log.Println("Connection pool closed")
	log.Println("Terminating PostgreSQL container...")

	err := s.psqlContainer.Terminate(ctx)
	s.Require().NoError(err)

	log.Println("PostgreSQL container terminated")
}

func TestSuite_Run(t *testing.T) {
	log.Println("Starting test suite execution")
	suite.Run(t, new(TestSuite))
	log.Println("Test suite execution completed")
}
