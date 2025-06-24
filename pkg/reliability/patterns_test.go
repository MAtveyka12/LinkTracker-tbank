package reliability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/reliability"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	circuitBreakerConfig := &reliability.CircuitBreakerConfig{
		SlidingWindowSize:        1,
		MinimumRequiredCalls:     1,
		FailureRateThreshold:     50,
		PermittedCallsInHalfOpen: 1,
		WaitDurationInOpenState:  1 * time.Second,
	}

	retryConfig := &reliability.RetryConfig{
		MaxRetries:      0,
		BackoffDuration: 0,
		RetryableCodes:  []int{},
	}

	rateLimitConfig := &reliability.RateLimitConfig{
		RequestsPerSecond: 1000,
		Burst:             1000,
	}

	client := reliability.NewHTTPClientWithReliability(
		2*time.Second,
		retryConfig,
		rateLimitConfig,
		circuitBreakerConfig,
	)

	for i := 0; i < 3; i++ {
		req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
		assert.NoError(t, err)

		resp, err := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		assert.Error(t, err)
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	assert.NoError(t, err)

	start := time.Now()
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	duration := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, duration, 2*time.Second+100*time.Millisecond, "Request should fail before client timeout when circuit breaker is open")
}

func TestRetryPattern(t *testing.T) {
	tests := []struct {
		name           string
		statusCodes    []int
		retryableCodes []int
		maxRetries     int
		shouldRetry    bool
		minCalls       int
		maxCalls       int
		expectError    bool
		expectStatus   int
	}{
		{
			name:           "Should retry on retryable status code",
			statusCodes:    []int{http.StatusInternalServerError, http.StatusOK},
			retryableCodes: []int{http.StatusInternalServerError},
			maxRetries:     3,
			shouldRetry:    true,
			minCalls:       1,
			maxCalls:       3,
			expectError:    false,
			expectStatus:   http.StatusOK,
		},
		{
			name:           "Should not retry on non-retryable status code",
			statusCodes:    []int{http.StatusBadRequest},
			retryableCodes: []int{http.StatusInternalServerError},
			maxRetries:     3,
			shouldRetry:    false,
			minCalls:       1,
			maxCalls:       2,
			expectError:    false,
			expectStatus:   http.StatusBadRequest,
		},
		{
			name:           "Should respect max retries limit",
			statusCodes:    []int{http.StatusInternalServerError, http.StatusInternalServerError, http.StatusInternalServerError},
			retryableCodes: []int{http.StatusInternalServerError},
			maxRetries:     2,
			shouldRetry:    true,
			minCalls:       1,
			maxCalls:       3,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				statusCode := tt.statusCodes[callCount%len(tt.statusCodes)]
				w.WriteHeader(statusCode)

				callCount++
			}))

			defer server.Close()

			retryConfig := &reliability.RetryConfig{
				MaxRetries:      tt.maxRetries,
				BackoffDuration: 10 * time.Millisecond,
				RetryableCodes:  tt.retryableCodes,
			}

			circuitBreakerConfig := &reliability.CircuitBreakerConfig{
				SlidingWindowSize:        1,
				MinimumRequiredCalls:     1,
				FailureRateThreshold:     50,
				PermittedCallsInHalfOpen: 1,
				WaitDurationInOpenState:  1 * time.Second,
			}

			rateLimitConfig := &reliability.RateLimitConfig{
				RequestsPerSecond: 1000,
				Burst:             1000,
			}

			client := reliability.NewHTTPClientWithReliability(
				500*time.Millisecond,
				retryConfig,
				rateLimitConfig,
				circuitBreakerConfig,
			)

			req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
			assert.NoError(t, err)

			resp, err := client.Do(req)

			if tt.expectError {
				assert.Error(t, err)

				if resp != nil {
					resp.Body.Close()
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)

				if resp != nil {
					defer resp.Body.Close()
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			}

			assert.GreaterOrEqual(t, callCount, tt.minCalls, "Should make at least minimum expected calls")
			assert.LessOrEqual(t, callCount, tt.maxCalls, "Should not exceed maximum expected calls")
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	retryConfig := &reliability.RetryConfig{
		MaxRetries:      3,
		BackoffDuration: 10 * time.Millisecond,
		RetryableCodes:  []int{http.StatusInternalServerError},
	}

	circuitBreakerConfig := &reliability.CircuitBreakerConfig{
		SlidingWindowSize:        1,
		MinimumRequiredCalls:     1,
		FailureRateThreshold:     50,
		PermittedCallsInHalfOpen: 1,
		WaitDurationInOpenState:  1 * time.Second,
	}

	rateLimitConfig := &reliability.RateLimitConfig{
		RequestsPerSecond: 1000,
		Burst:             1000,
	}

	client := reliability.NewHTTPClientWithReliability(
		500*time.Millisecond,
		retryConfig,
		rateLimitConfig,
		circuitBreakerConfig,
	)

	req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	assert.NoError(t, err)

	start := time.Now()
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, callCount, 1, "Should have made at least 1 call")
	assert.LessOrEqual(t, callCount, 3, "Should not have made more than 3 calls")
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond, "Should have waited at least 10ms between retries")
}
