package stackoverflow_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sof "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/stackoverflow"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client/stackoverflow"
)

func TestGetLatestAnswers(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение ответов", func(t *testing.T) {
		expectedAnswers := []sof.Answer{
			{AnswerID: 1, Body: "This is an answer"},
			{AnswerID: 2, Body: "Another answer"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasPrefix(r.URL.Path, "/questions/"))
			assert.Contains(t, r.URL.RawQuery, "order=desc")
			assert.Contains(t, r.URL.RawQuery, "sort=creation")
			assert.Contains(t, r.URL.RawQuery, "site=stackoverflow")
			assert.Contains(t, r.URL.RawQuery, "filter=withbody")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": expectedAnswers,
			})
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		answers, err := client.GetLatestAnswers(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, expectedAnswers, answers)
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		client := &stackoverflow.Client{
			BaseURL:    "http://invalid_url",
			HTTPClient: &http.Client{Timeout: time.Millisecond},
		}

		_, err := client.GetLatestAnswers(ctx, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request execution error")
	})

	t.Run("некорректный код ответа", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error_message": "Access denied",
			})
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestAnswers(ctx, 123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error API Stack Overflow: Access denied")
	})

	t.Run("ошибка декодирования ответа", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestAnswers(ctx, 123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding the response")
	})
}

func TestGetLatestComments(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение комментариев", func(t *testing.T) {
		expectedComments := []sof.Comment{
			{CommentID: 1, Body: "This is a comment"},
			{CommentID: 2, Body: "Another comment"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasPrefix(r.URL.Path, "/questions/"))
			assert.Contains(t, r.URL.RawQuery, "order=desc")
			assert.Contains(t, r.URL.RawQuery, "sort=creation")
			assert.Contains(t, r.URL.RawQuery, "site=stackoverflow")
			assert.Contains(t, r.URL.RawQuery, "filter=withbody")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": expectedComments,
			})
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		comments, err := client.GetLatestComments(ctx, 456)
		require.NoError(t, err)
		assert.Equal(t, expectedComments, comments)
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		client := &stackoverflow.Client{
			BaseURL:    "http://invalid_url",
			HTTPClient: &http.Client{Timeout: time.Millisecond},
		}

		_, err := client.GetLatestComments(ctx, 456)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request execution error")
	})

	t.Run("некорректный код ответа", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error_message": "Access denied",
			})
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestComments(ctx, 456)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error API Stack Overflow: Access denied")
	})

	t.Run("ошибка декодирования ответа", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &stackoverflow.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestComments(ctx, 456)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding the response")
	})
}

func TestGetType(t *testing.T) {
	client := &stackoverflow.Client{}
	assert.Equal(t, clients.StackOverflowType, client.GetType())
}
