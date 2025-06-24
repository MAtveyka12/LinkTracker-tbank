package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"golang.org/x/time/rate"

	// Register PostgreSQL driver.
	_ "github.com/lib/pq"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/scheduler"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/service"
	prodDomain "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client/stackoverflow"
	httpHandler "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/http"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer"
	httpProd "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer/http"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
	pgx "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/pgxpool"
	dbsql "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/sql"
	httpserver "github.com/es-debug/backend-academy-2024-go-template/pkg/http_server"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware"
)

type App struct {
	logger     *slog.Logger
	closer     *Closer
	httpServer *httpserver.HTTPServer
	producer   producer.ProdKafkaOrHTTP
}

func New() (*App, error) {
	logger := setupLogger()
	cfg, err := loadConfig(logger)

	if err != nil {
		return nil, err
	}

	closer := NewCloser()
	pool, db, err := setupDatabase(cfg, logger, closer)

	if err != nil {
		return nil, err
	}

	updateChan := make(chan prodDomain.LinkUpdate)
	prod, err := setupProducer(cfg, logger, updateChan)

	if err != nil {
		return nil, err
	}

	closer.Add(func(_ context.Context) error {
		logger.Info("Stopping Producer...!")
		close(updateChan)

		if prod != nil {
			prod.Stop()
		}

		return nil
	})

	txManager, linkRepo, chatRepo := setupRepositories(cfg, pool, db, logger)
	clnt := []clients.Client{
		github.NewGitHubClient(cfg),
		stackoverflow.NewStackOverflowClient(cfg),
	}

	logger.Debug("Clients initialized",
		slog.Int("count", len(clnt)),
		slog.Any("clients", []string{"github", "stackoverflow"}),
	)

	sched := scheduler.NewScheduler(clnt, logger, updateChan, prod, cfg)
	linkService := service.NewLinkService(linkRepo, txManager, logger, sched)
	chatService := service.NewChatService(chatRepo, linkRepo, txManager, logger)
	serv := service.NewService(chatService, linkService)
	srv := setupHTTPServer(cfg, serv, logger)

	return &App{
		logger:     logger,
		httpServer: srv,
		closer:     closer,
		producer:   prod,
	}, nil
}

func (s *App) Start(ctx context.Context) error {
	defer func() {
		s.logger.Info("Closing resources...")

		if err := s.closer.Close(ctx); err != nil {
			s.logger.Error("Failed to close all connections", slog.String("error", err.Error()))
		}

		s.logger.Info("Closed all connections!")
	}()

	s.logger.Info("Starting the server...")

	err := s.httpServer.Start(ctx)
	if err != nil {
		return fmt.Errorf("server startup error: %s", err.Error())
	}

	return nil
}

func setupLogger() *slog.Logger {
	return slog.New(tint.NewHandler(os.Stdout, nil))
}

func loadConfig(logger *slog.Logger) (*config.Config, error) {
	cfg, err := config.LoadConfig("config/config.yaml", ".env")
	if err != nil {
		logger.Error("Failed to load the config", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("Config loaded!")

	return cfg, nil
}

func setupDatabase(cfg *config.Config, logger *slog.Logger, closer *Closer) (*pgxpool.Pool, *sql.DB, error) {
	pool, err := cfg.Database.GetPgxConnection(context.TODO())
	if err != nil {
		logger.Error("Failed to connect to DB", slog.String("error", err.Error()))
		return nil, nil, err
	}

	logger.Info("Connected to Pgx Pool!")

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping DB", slog.String("db", cfg.Database.GetDSN()), slog.String("error", err.Error()))
		return nil, nil, err
	}

	closer.Add(func(context.Context) error {
		pool.Close()
		return nil
	})

	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		return nil, nil, err
	}

	closer.Add(func(context.Context) error {
		db.Close()
		return nil
	})

	return pool, db, nil
}

func setupProducer(cfg *config.Config, logger *slog.Logger, updateChan chan prodDomain.LinkUpdate) (producer.ProdKafkaOrHTTP, error) {
	switch cfg.MessageTransport {
	case "kafka":
		saramaConfig := sarama.NewConfig()

		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Return.Errors = true

		kafkaProducer, err := kafka.NewKafkaProducer(cfg.Kafka.Addresses, logger, updateChan, cfg.Kafka.Topic, cfg.Kafka.DLQTopic, saramaConfig)
		if err != nil {
			logger.Error("Failed to create a new Kafka producer", slog.String("error", err.Error()))
			return nil, err
		}

		kafkaProducer.Run()

		logger.Info("Successfully created and started Kafka producer")

		return kafkaProducer, nil
	case "http":
		httpProducer := httpProd.NewHTTPProducer()

		logger.Info("Successfully created and started HTTP producer")

		return httpProducer, nil
	default:
		logger.Error("Unknown message transport in config", slog.String("message_transport", cfg.MessageTransport))
		return nil, fmt.Errorf("error when creating producer, it must be either kafka or http")
	}
}

func setupRepositories(cfg *config.Config, pool *pgxpool.Pool, db *sql.DB, logger *slog.Logger) (
	repository.TxManager,
	repository.LinkRepository,
	repository.ChatRepository,
) {
	switch cfg.AccessType {
	case "sql":
		ctxManager := dbsql.NewCtxManager(db)
		txManager := dbsql.NewTxManager(db, logger, ctxManager)

		logger.Debug("Transaction manager initialized",
			slog.Bool("with_logging", true),
		)

		linkRepo := dbsql.NewLinkRepository(ctxManager)
		chatRepo := dbsql.NewChatRepository(ctxManager)

		logger.Info("Repositories initialized with Sql")

		return txManager, linkRepo, chatRepo
	case "orm":
		ctxManager := pgx.NewCtxManager(pool)
		txManager := pgx.NewTxManager(pool, logger, ctxManager)

		logger.Debug("Transaction manager initialized",
			slog.Bool("with_logging", true),
		)

		linkRepo := pgx.NewLinkRepository(ctxManager)
		chatRepo := pgx.NewChatRepository(ctxManager)

		logger.Info("Repositories initialized with Pgx")

		return txManager, linkRepo, chatRepo
	default:
		logger.Error("Unknown access type in config", slog.String("access_type", cfg.AccessType))
		return nil, nil, nil
	}
}

func setupHTTPServer(cfg *config.Config, services service.Service, logger *slog.Logger) *httpserver.HTTPServer {
	handler := httpHandler.NewHandler(services)
	router := gin.Default()

	rateLimiter := middleware.NewIPRateLimiter(
		rate.Limit(cfg.Reliability.Middleware.RateLimit.RequestsPerSecond),
		cfg.Reliability.Middleware.RateLimit.Burst,
		cfg.Reliability.Middleware.RateLimit.TTL,
		cfg.Reliability.Middleware.RateLimit.Cleanup,
	)

	router.Use(func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}

		limiter := rateLimiter.GetLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	})

	router.POST("/tg-chat/:id", handler.HandlerRegisterChat)
	router.DELETE("/tg-chat/:id", handler.HandlerDeleteChat)
	router.GET("/links", handler.HandlerGetAllLinks)
	router.POST("/links", handler.HandlerAddLink)
	router.DELETE("/links", handler.HandlerDeleteLink)

	httpServerConfig := &httpserver.Config{
		Host:              cfg.Scrapper.Host,
		Port:              cfg.Scrapper.Port,
		StartMsg:          "Server started!",
		Handler:           router,
		ReadTimeout:       cfg.Reliability.HTTP.Timeout,
		WriteTimeout:      cfg.Reliability.HTTP.Timeout,
		ReadHeaderTimeout: cfg.Reliability.HTTP.Timeout,
		ShutdownTimeout:   5 * time.Second,
	}

	return httpserver.NewHTTPServer(logger, httpServerConfig)
}
