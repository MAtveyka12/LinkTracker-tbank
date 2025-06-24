package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Database struct {
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	Name              string `yaml:"name"`
	SSLMode           string `yaml:"sslmode"`
	MaxConns          int32  `yaml:"max_connections"`
	MinConns          int32  `yaml:"min_connections"`
	MaxLifeTime       int    `yaml:"max_lifetime"`
	MaxIdleTime       int    `yaml:"max_idle_time"`
	HealthCheckPeriod int    `yaml:"health_check_period"`

	User     string `yaml:"-"`
	Password string `yaml:"-"`
}

type Scrapper struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Bot struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Kafka struct {
	Addresses []string
	Topic     string
	DLQTopic  string
	Group     string
}

type Redis struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type CircuitBreakerSettings struct {
	SlidingWindowSize        int           `yaml:"sliding_window_size"`
	MinimumRequiredCalls     uint32        `yaml:"minimum_required_calls"`
	FailureRateThreshold     float64       `yaml:"failure_rate_threshold"`
	PermittedCallsInHalfOpen uint32        `yaml:"permitted_calls_in_half_open"`
	WaitDurationInOpenState  time.Duration `yaml:"wait_duration_in_open_state"`
}

type HTTPReliability struct {
	Timeout time.Duration `yaml:"timeout"`
	Retry   struct {
		MaxRetries      int           `yaml:"max_retries"`
		BackoffDuration time.Duration `yaml:"backoff_duration"`
		RetryableCodes  []int         `yaml:"retryable_codes"`
	} `yaml:"retry"`
	RateLimit struct {
		RequestsPerSecond float64 `yaml:"requests_per_second"`
		Burst             int     `yaml:"burst"`
	} `yaml:"rate_limit"`
	CircuitBreaker CircuitBreakerSettings `yaml:"circuit_breaker"`
}

type MiddlewareReliability struct {
	RateLimit struct {
		RequestsPerSecond float64       `yaml:"requests_per_second"`
		Burst             int           `yaml:"burst"`
		TTL               time.Duration `yaml:"ttl"`
		Cleanup           time.Duration `yaml:"cleanup"`
	} `yaml:"rate_limit"`
}

type Reliability struct {
	HTTP       HTTPReliability       `yaml:"http"`
	Middleware MiddlewareReliability `yaml:"middleware"`
}

type Config struct {
	Database         Database `yaml:"database"`
	Scrapper         Scrapper `yaml:"scrapper_server"`
	Bot              Bot      `yaml:"bot_server"`
	Redis            Redis    `yaml:"redis"`
	GitHubToken      string   `yaml:"-"`
	TelegramToken    string   `yaml:"-"`
	AccessType       string   `yaml:"access_type"`
	MessageTransport string   `yaml:"message_transport"`
	Kafka            Kafka
	Reliability      Reliability `yaml:"reliability"`
}

func (d *Database) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (d *Database) GetPgxConnection(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := d.GetDSN()
	cfg, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("DSN parsing error for PostgreSQL: %w", err)
	}

	if d.MaxConns > 0 {
		cfg.MaxConns = d.MaxConns
	}

	if d.MinConns > 0 {
		cfg.MinConns = d.MinConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to create a pool of connections with PostgreSQL: %w", err)
	}

	return pool, nil
}

func (s *Scrapper) URL() string {
	return fmt.Sprintf("http://%s:%s", s.Host, s.Port)
}

func (r *Redis) URL() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func LoadConfig(configPath, envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(file, &config)

	if err != nil {
		return nil, err
	}

	if config.MessageTransport != "kafka" && config.MessageTransport != "http" {
		return nil, fmt.Errorf("invalid message transport: %s, must be 'kafka' or 'http'", config.MessageTransport)
	}

	config.GitHubToken = os.Getenv("GITHUB_TOKEN")
	config.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	config.Database.User = os.Getenv("DB_USER")
	config.Database.Password = os.Getenv("DB_PASSWORD")
	config.Kafka.Addresses = []string{os.Getenv("KAFKA_BROKERS")}
	config.Kafka.Topic = os.Getenv("KAFKA_TOPIC")
	config.Kafka.DLQTopic = os.Getenv("KAFKA_DLQ")
	config.Kafka.Group = os.Getenv("KAFKA_CONSUMER_GROUP")

	return &config, nil
}
