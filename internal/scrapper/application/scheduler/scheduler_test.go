package scheduler_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/scheduler"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	mockClients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client/mock"
	mockProducer "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer/mock"
)

func TestRemoveLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	s := scheduler.NewScheduler(nil, logger, nil, mockProd, cfg)

	_ = s.AddLink("https://github.com/user/repo", 1, 1)
	err := s.RemoveLink("https://github.com/user/repo")
	assert.NoError(t, err)
}

func TestAddLink_InvalidFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	sched := scheduler.NewScheduler(nil, mockLogger, make(chan producer.LinkUpdate), mockProd, cfg)

	err := sched.AddLink("invalid_link", 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "incorrect link format", err.Error())
}

func TestAddLink_ValidGitHub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockClients.NewMockClient(ctrl)
	mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	sched := scheduler.NewScheduler([]clients.Client{mockClient}, mockLogger, make(chan producer.LinkUpdate), mockProd, cfg)

	mockClient.EXPECT().GetType().Return(clients.GitHubType).AnyTimes()

	err := sched.AddLink("https://github.com/user/repo", 1, 1)
	assert.NoError(t, err)
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	sched := scheduler.NewScheduler(nil, mockLogger, make(chan producer.LinkUpdate), mockProd, cfg)

	_ = sched.Stop()
}

func TestMonitorRepo_NoGitHubClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	s := scheduler.NewScheduler(nil, logger, nil, mockProd, cfg)

	err := s.MonitorRepo(context.Background(), "user", "repo", "https://github.com/user/repo", 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "the GitHub client is not configured", err.Error())
}

func TestHandleCommits_NoCommits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockClients.NewMockClient(ctrl)
	mockClient.EXPECT().GetLatestCommits(gomock.Any(), "user", "repo").Return(nil, errors.New("ошибка запроса"))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	s := scheduler.NewScheduler([]clients.Client{mockClient}, logger, nil, mockProd, cfg)

	err := s.HandleCommits(context.Background(), mockClient, "user", "repo", "https://github.com/user/repo", 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error when receiving Commits")
}

func TestHandleCommits_SendsUpdatesOnlyToSubscribers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockClients.NewMockClient(ctrl)
	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}
	s := scheduler.NewScheduler([]clients.Client{mockClient}, nil, make(chan producer.LinkUpdate, 1), mockProd, cfg)

	link := "https://github.com/user/repo"
	userSubscribed := 101
	linkID := 1

	mockClient.EXPECT().GetLatestCommits(gomock.Any(), "user", "repo").Return([]github.Commit{
		{
			SHA: "",
			Commit: struct {
				Message   string        `json:"message"`
				Author    github.Author `json:"author"`
				Committer github.Author `json:"committer"`
			}{
				Message: "New commit",
			},
			URL: "",
		},
	}, nil)

	mockProd.EXPECT().
		SendUpdate(producer.LinkUpdate{
			ID:          1,
			URL:         "https://github.com/user/repo",
			Description: "New commit",
			UserID:      101,
			Type:        "Commit",
		}).
		Return(nil)

	err := s.HandleCommits(context.Background(), mockClient, "user", "repo", link, userSubscribed, linkID)

	assert.NoError(t, err, "Обработчик коммитов не должен возвращать ошибку")
}

func TestAddLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockClients.NewMockClient(ctrl)
	mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockClient.EXPECT().GetType().Return(clients.GitHubType).AnyTimes()

	mockProd := mockProducer.NewMockProdKafkaOrHTTP(ctrl)
	cfg := &config.Config{MessageTransport: "http"}

	sched := scheduler.NewScheduler([]clients.Client{mockClient}, mockLogger, make(chan producer.LinkUpdate), mockProd, cfg)

	err := sched.AddLink("https://github.com/user/repo", 1, 1)
	assert.NoError(t, err)

	err = sched.AddLink("https://stackoverflow.com/questions/12345678/question-title", 2, 2)
	assert.NoError(t, err)

	err = sched.AddLink("https://example.com/user/repo", 3, 3)
	assert.Error(t, err)
}
