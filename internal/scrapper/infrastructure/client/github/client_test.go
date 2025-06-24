package github_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	gh "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/github"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client/github"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/reliability/mock"
)

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

func TestGetLatestCommits(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	t.Run("успешное получение коммитов", func(t *testing.T) {
		expectedCommits := []gh.Commit{
			{
				SHA: "123456",
				Commit: struct {
					Message   string    `json:"message"`
					Author    gh.Author `json:"author"`
					Committer gh.Author `json:"committer"`
				}{
					Message: "Initial commit",
				},
				URL: "",
			},
			{
				SHA: "abcdef",
				Commit: struct {
					Message   string    `json:"message"`
					Author    gh.Author `json:"author"`
					Committer gh.Author `json:"committer"`
				}{
					Message: "Added new feature",
				},
				URL: "",
			},
		}

		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))
			assert.True(t, strings.HasPrefix(req.URL.Path, "/repos/"))
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(mustMarshal(expectedCommits))),
				Header:     make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
			Token:      "test-token",
		}

		commits, err := client.GetLatestCommits(ctx, "owner", "repo")
		require.NoError(t, err)
		assert.Equal(t, expectedCommits, commits)
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("connection refused"))

		client := &github.Client{
			BaseURL:    "http://invalid_url",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "owner", "repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request execution error")
	})

	t.Run("некорректный код ответа", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body: io.NopCloser(bytes.NewReader(mustMarshal(map[string]interface{}{
					"error_message": "Access denied",
				}))),
				Header: make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "owner", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error API GitHub: Access denied (code: 403)")
	})

	t.Run("ошибка декодирования ответа", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				Header:     make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "owner", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding commits")
	})

	t.Run("обработка ошибки 500 от GitHub API (с JSON-ошибкой)", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body: io.NopCloser(bytes.NewReader(mustMarshal(map[string]interface{}{
					"error_message": "internal server error",
				}))),
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "user", "repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error API GitHub: internal server error (code: 500)")
	})

	t.Run("некорректное тело ответа", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				Header:     make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "user", "repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding commits")
	})

	t.Run("обработка ошибки 404 (не найдено)", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body: io.NopCloser(bytes.NewReader(mustMarshal(map[string]interface{}{
					"error_message": "репозиторий не найден",
				}))),
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestCommits(ctx, "user", "repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error API GitHub: репозиторий не найден (code: 404)")
	})
}

func TestGetLatestPRs(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	t.Run("успешное получение PR", func(t *testing.T) {
		expectedPRs := []gh.PullRequest{
			{ID: 1, Title: "Fix bug #123"},
			{ID: 2, Title: "Add new feature"},
		}

		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))
			assert.True(t, strings.HasPrefix(req.URL.Path, "/repos/"))
			assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(mustMarshal(expectedPRs))),
				Header:     make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
			Token:      "test-token",
		}

		prs, err := client.GetLatestPRs(ctx, "owner", "repo")
		require.NoError(t, err)
		assert.Equal(t, expectedPRs, prs)
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("connection refused"))

		client := &github.Client{
			BaseURL:    "http://invalid_url",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestPRs(ctx, "owner", "repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request execution error")
	})

	t.Run("некорректный код ответа", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body: io.NopCloser(bytes.NewReader(mustMarshal(map[string]interface{}{
					"error_message": "Access denied",
				}))),
				Header: make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestPRs(ctx, "owner", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error API GitHub: Access denied (code: 403)")
	})

	t.Run("ошибка декодирования ответа", func(t *testing.T) {
		mockClient := mock.NewMockHTTPClient(ctrl)
		mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "application/vnd.github.v3+json", req.Header.Get("Accept"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				Header:     make(http.Header),
			}, nil
		})

		client := &github.Client{
			BaseURL:    "http://api.github.com",
			HTTPClient: mockClient,
		}

		_, err := client.GetLatestPRs(ctx, "owner", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding error PR")
	})
}

func TestGetType(t *testing.T) {
	client := &github.Client{}
	assert.Equal(t, clients.GitHubType, client.GetType())
}
