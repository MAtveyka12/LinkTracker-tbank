package stackoverflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/stackoverflow"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/reliability"
)

type Client struct {
	BaseURL    string
	HTTPClient reliability.HTTPClient
}

func NewStackOverflowClient(cfg *config.Config) clients.Client {
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

	httpClient := reliability.NewHTTPClientWithReliability(
		cfg.Reliability.HTTP.Timeout,
		retryConfig,
		rateLimitConfig,
		circuitBreakerConfig,
	)

	return &Client{
		BaseURL:    "https://api.stackexchange.com/2.3",
		HTTPClient: httpClient,
	}
}

func (c *Client) GetType() clients.ClientType {
	return clients.StackOverflowType
}

type itemsResponse[T any] struct {
	Items []T `json:"items"`
}

func (c *Client) GetLatestAnswers(ctx context.Context, questionID int) ([]stackoverflow.Answer, error) {
	url := fmt.Sprintf("%s/questions/%d/answers?order=desc&sort=creation&site=stackoverflow&filter=withbody", c.BaseURL, questionID)
	return fetchItems[stackoverflow.Answer](ctx, c, url)
}

func (c *Client) GetLatestComments(ctx context.Context, questionID int) ([]stackoverflow.Comment, error) {
	url := fmt.Sprintf("%s/questions/%d/comments?order=desc&sort=creation&site=stackoverflow&filter=withbody", c.BaseURL, questionID)
	return fetchItems[stackoverflow.Comment](ctx, c, url)
}

func fetchItems[T any](ctx context.Context, client *Client, url string) ([]T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var soError struct {
			ErrorMessage string `json:"error_message"`
		}

		err := json.NewDecoder(resp.Body).Decode(&soError)
		if err != nil {
			return nil, fmt.Errorf("data decoding error: %s", err.Error())
		}

		return nil, fmt.Errorf("error API Stack Overflow: %s", soError.ErrorMessage)
	}

	var result itemsResponse[T]

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding the response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetLatestCommits(_ context.Context, _, _ string) ([]github.Commit, error) {
	return nil, nil
}

func (c *Client) GetLatestPRs(_ context.Context, _, _ string) ([]github.PullRequest, error) {
	return nil, nil
}
