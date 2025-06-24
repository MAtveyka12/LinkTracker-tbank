package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/application"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/adapter/telegram"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/cache"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/kafka"
	httpclient "github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/scrapper/http_client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/storage"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/reliability"
)

type App struct {
	config   *config.Config
	logger   *slog.Logger
	bot      *telegram.TgBot
	server   *http.Server
	consumer *kafka.Consumer
}

func New() (*App, error) {
	logger := slog.Default()
	cfg, err := config.LoadConfig("config/config.yaml", ".env")

	if err != nil {
		return nil, fmt.Errorf("configuration loading error: %s", err.Error())
	}

	addresScrapper := cfg.Scrapper.URL()
	repo := storage.NewRepository()
	retryConfig, rateLimitConfig, circuitBreakerConfig := reliabilityStart(cfg)
	scrpr := httpclient.NewScraperHTTPClient(addresScrapper, retryConfig, rateLimitConfig, circuitBreakerConfig)
	addresRedis := cfg.Redis.URL()
	redisCache := cache.NewRedisCache(addresRedis, 0, "", time.Minute)
	tgBot, err := telegram.NewBot(cfg, repo, scrpr, logger, redisCache)

	if err != nil {
		return nil, fmt.Errorf("the bot has not been created: %s", err.Error())
	}

	application.SetupRouter(tgBot)

	botAddress := fmt.Sprintf("%s:%s", cfg.Bot.Host, cfg.Bot.Port)

	server := &http.Server{
		Addr:         botAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	updateChan := make(chan sarama.ConsumerMessage)
	saramaConfig := sarama.NewConfig()

	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Version = sarama.V2_8_0_0

	kafkaConsumer, err := kafka.NewKafkaConsumer(
		cfg.Kafka.Addresses,
		cfg.Kafka.Topic,
		cfg.Kafka.Group,
		updateChan,
		logger,
		saramaConfig,
	)

	if err != nil {
		logger.Error("Failed to create a new Sarama consumer", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("Successfully created a Kafka consumer")

	return &App{
		config:   cfg,
		logger:   logger,
		bot:      tgBot,
		server:   server,
		consumer: kafkaConsumer,
	}, nil
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.logger.Info("The bot is running...")
	go a.bot.Bot.Start()

	switch a.config.MessageTransport {
	case "kafka":
		err := a.consumer.Run(ctx)
		if err != nil {
			a.logger.Error("Failed to run Kafka consumer", slog.String("error", err.Error()))
			return err
		}

		go func() {
			for msg := range a.consumer.Messages() {
				a.logger.Info("Received Kafka message",
					slog.String("topic", msg.Topic),
					slog.String("key", string(msg.Key)),
					slog.String("value", string(msg.Value)),
				)

				if err := a.bot.KafkaMessageHandler(&msg); err != nil {
					a.logger.Error("Failed to handle Kafka message", slog.String("error", err.Error()))
				}
			}
		}()
	case "http":
		go func() {
			http.HandleFunc("/updates", telegram.UpdatesHandler(a.bot))

			a.logger.Info("The HTTP server for processing updates on port 7070 is running.")

			err := a.server.ListenAndServe()

			if err != nil && err != http.ErrServerClosed {
				a.logger.Error(fmt.Sprintf("HTTP Server startup error: %s", err.Error()))
			}
		}()
	default:
		a.logger.Error("Unknown message transport in config", slog.String("message_transport", a.config.MessageTransport))
		return fmt.Errorf("error when creating producer, it must be either kafka or http")
	}

	<-stop

	cancel()
	a.consumer.Stop()
	a.consumer.Wait()

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	a.logger.Info("Stopping the bot server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error(fmt.Sprintf("Error when shutting down the bot server: %s", err.Error()))
		return err
	}

	a.logger.Info("The bot's server is stopped correctly")

	return nil
}

func reliabilityStart(cfg *config.Config) (
	*reliability.RetryConfig,
	*reliability.RateLimitConfig,
	*reliability.CircuitBreakerConfig,
) {
	retryConfig := &reliability.RetryConfig{
		MaxRetries:      cfg.Reliability.HTTP.Retry.MaxRetries,
		BackoffDuration: cfg.Reliability.HTTP.Retry.BackoffDuration,
		RetryableCodes:  cfg.Reliability.HTTP.Retry.RetryableCodes,
	}

	rateLimitConfig := &reliability.RateLimitConfig{
		RequestsPerSecond: cfg.Reliability.HTTP.RateLimit.RequestsPerSecond,
		Burst:             cfg.Reliability.HTTP.RateLimit.Burst,
	}

	circuitBreakerConfig := &reliability.CircuitBreakerConfig{
		SlidingWindowSize:        cfg.Reliability.HTTP.CircuitBreaker.SlidingWindowSize,
		MinimumRequiredCalls:     cfg.Reliability.HTTP.CircuitBreaker.MinimumRequiredCalls,
		FailureRateThreshold:     cfg.Reliability.HTTP.CircuitBreaker.FailureRateThreshold,
		PermittedCallsInHalfOpen: cfg.Reliability.HTTP.CircuitBreaker.PermittedCallsInHalfOpen,
		WaitDurationInOpenState:  cfg.Reliability.HTTP.CircuitBreaker.WaitDurationInOpenState,
	}

	return retryConfig, rateLimitConfig, circuitBreakerConfig
}
