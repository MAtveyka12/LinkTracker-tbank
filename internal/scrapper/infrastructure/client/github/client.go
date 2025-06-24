package github

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
	Token      string
}

func NewGitHubClient(cfg *config.Config) clients.Client {
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
		BaseURL:    "https://api.github.com",
		HTTPClient: httpClient,
		Token:      cfg.GitHubToken,
	}
}

func (c *Client) GetType() clients.ClientType {
	return clients.GitHubType
}

func (c *Client) GetLatestCommits(ctx context.Context, owner, repo string) ([]github.Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits", c.BaseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("request creation error: %s", err.Error())
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var ghError struct {
			ErrorMessage string `json:"error_message"`
		}

		err := json.NewDecoder(resp.Body).Decode(&ghError)
		if err != nil {
			return nil, fmt.Errorf("data decoding error: %s", err.Error())
		}

		return nil, fmt.Errorf("error API GitHub: %s (code: %d)", ghError.ErrorMessage, resp.StatusCode)
	}

	var commits []github.Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("error decoding commits: %s", err.Error())
	}

	return commits, nil
}

func (c *Client) GetLatestPRs(ctx context.Context, owner, repo string) ([]github.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.BaseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request execution error: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var ghError struct {
			ErrorMessage string `json:"error_message"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&ghError); err != nil {
			return nil, fmt.Errorf("data decoding error: %s", err.Error())
		}

		return nil, fmt.Errorf("error API GitHub: %s (code: %d)", ghError.ErrorMessage, resp.StatusCode)
	}

	var prs []github.PullRequest

	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("decoding error PR: %s", err.Error())
	}

	return prs, nil
}

func (c *Client) GetLatestAnswers(_ context.Context, _ int) ([]stackoverflow.Answer, error) {
	return nil, nil
}

func (c *Client) GetLatestComments(_ context.Context, _ int) ([]stackoverflow.Comment, error) {
	return nil, nil
}
