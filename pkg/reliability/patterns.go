package reliability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RetryConfig struct {
	MaxRetries      int
	BackoffDuration time.Duration
	RetryableCodes  []int
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

type CircuitBreakerConfig struct {
	SlidingWindowSize        int
	MinimumRequiredCalls     uint32
	FailureRateThreshold     float64
	PermittedCallsInHalfOpen uint32
	WaitDurationInOpenState  time.Duration
}

type HTTPClientWithReliability struct {
	client         *http.Client
	retryConfig    *RetryConfig
	rateLimiter    *rate.Limiter
	circuitBreaker *gobreaker.CircuitBreaker
}

func NewHTTPClientWithReliability(
	timeout time.Duration,
	retryConfig *RetryConfig,
	rateLimitConfig *RateLimitConfig,
	circuitBreakerConfig *CircuitBreakerConfig,
) HTTPClient {
	client := &http.Client{
		Timeout: timeout,
	}

	limiter := rate.NewLimiter(rate.Limit(rateLimitConfig.RequestsPerSecond), rateLimitConfig.Burst)

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "http-client",
		MaxRequests: circuitBreakerConfig.PermittedCallsInHalfOpen,
		Interval:    time.Duration(circuitBreakerConfig.SlidingWindowSize) * time.Second,
		Timeout:     circuitBreakerConfig.WaitDurationInOpenState,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < circuitBreakerConfig.MinimumRequiredCalls {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= circuitBreakerConfig.FailureRateThreshold/100
		},
		OnStateChange: func(_ string, _ gobreaker.State, _ gobreaker.State) {},
	})

	return &HTTPClientWithReliability{
		client:         client,
		retryConfig:    retryConfig,
		rateLimiter:    limiter,
		circuitBreaker: cb,
	}
}

func (c *HTTPClientWithReliability) Do(req *http.Request) (*http.Response, error) {
	if err := c.rateLimiter.Wait(req.Context()); err != nil {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       http.NoBody,
		}, nil
	}

	response, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		var resp *http.Response

		var lastErr error

		for i := 0; i <= c.retryConfig.MaxRetries; i++ {
			resp, lastErr = c.client.Do(req)
			if lastErr == nil {
				shouldRetry := false

				for _, code := range c.retryConfig.RetryableCodes {
					if resp.StatusCode == code {
						shouldRetry = true
						break
					}
				}

				if !shouldRetry {
					return resp, nil
				}

				resp.Body.Close()
			}

			if i < c.retryConfig.MaxRetries {
				time.Sleep(c.retryConfig.BackoffDuration)
				continue
			}
		}

		if lastErr == nil {
			return nil, fmt.Errorf("circuit breaker is open")
		}

		return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
	})

	if err != nil {
		return nil, err
	}

	return response.(*http.Response), nil
}

func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
