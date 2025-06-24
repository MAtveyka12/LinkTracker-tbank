package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	mock_scheduler "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/scheduler/mock"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/mock"
)

const (
	link = "https://example.com"
)

func TestRegisterChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mock.NewMockChatRepository(ctrl)
	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewChatService(mockChatRepo, mockLinkRepo, mockTxManager, logger)
	ctx := context.Background()
	userID := 123

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()

	mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return([]string{"link1"}, nil)

	err := serv.RegisterChat(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, "the chat has already been registered", err.Error())

	mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return(nil, errors.New("not found"))
	mockChatRepo.EXPECT().RegisterChat(ctx, userID).Return(nil)

	err = serv.RegisterChat(ctx, userID)
	assert.NoError(t, err)
}

func TestDeleteChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mock.NewMockChatRepository(ctrl)
	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewChatService(mockChatRepo, mockLinkRepo, mockTxManager, logger)
	ctx := context.Background()
	userID := 123

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return([]string{"link1", "link2"}, nil)
	mockLinkRepo.EXPECT().DeleteAllLinks(ctx, userID).Return(nil)

	err := serv.DeleteChat(ctx, userID)
	assert.NoError(t, err)
}

func TestAddLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockScheduler := mock_scheduler.NewMockScheduler(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewLinkService(mockLinkRepo, mockTxManager, logger, mockScheduler)
	ctx := context.Background()
	userID := 123

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)

	mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return([]string{}, nil)
	mockLinkRepo.EXPECT().AddLink(ctx, userID, link).Return(1, nil)
	mockScheduler.EXPECT().AddLink(link, userID, 1).Return(nil)

	linkID, err := serv.AddLink(ctx, userID, link)
	assert.NoError(t, err)
	assert.Equal(t, 1, linkID)

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)

	mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return([]string{link}, nil)
	_, err = serv.AddLink(ctx, userID, link)
	assert.Error(t, err)
	assert.Equal(t, "Link already exists", err.Error())
}

func TestDeleteLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockScheduler := mock_scheduler.NewMockScheduler(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewLinkService(mockLinkRepo, mockTxManager, logger, mockScheduler)
	ctx := context.Background()
	userID := 123
	linkID := 1

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			mockLinkRepo.EXPECT().
				GetLinkIDByURL(ctx, userID, link).
				Return(linkID, nil)

			mockLinkRepo.EXPECT().
				DeleteLink(ctx, userID, linkID, link).
				Return(nil)

			mockScheduler.EXPECT().
				RemoveLink(link).
				Return(nil)

			return fn(ctx)
		}).Times(1)

	err := serv.DeleteLink(ctx, userID, link)
	assert.NoError(t, err)
}

func TestGetAllLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockScheduler := mock_scheduler.NewMockScheduler(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewLinkService(mockLinkRepo, mockTxManager, logger, mockScheduler)
	ctx := context.Background()
	userID := 123

	expectedStrings := []string{"https://example.com", "https://test.com"}
	expectedLinks := []service.Link{
		{URL: expectedStrings[0]},
		{URL: expectedStrings[1]},
	}

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return(expectedStrings, nil)
			return fn(ctx)
		}).Times(1)

	links, err := serv.GetAllLinks(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedLinks, links)

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return(nil, errors.New("database error"))
			return fn(ctx)
		}).Times(1)

	links, err = serv.GetAllLinks(ctx, userID)
	assert.Error(t, err)
	assert.Nil(t, links)
}

func TestAddDuplicateLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinkRepo := mock.NewMockLinkRepository(ctrl)
	mockScheduler := mock_scheduler.NewMockScheduler(ctrl)
	mockTxManager := mock.NewMockTxManager(ctrl)
	logger := slog.Default()

	serv := service.NewLinkService(mockLinkRepo, mockTxManager, logger, mockScheduler)
	ctx := context.Background()
	userID := 123

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			mockLinkRepo.EXPECT().GetAllLinks(ctx, userID).Return([]string{link}, nil)
			return fn(ctx)
		}).Times(1)

	_, err := serv.AddLink(ctx, userID, link)
	assert.Error(t, err)
	assert.Equal(t, "Link already exists", err.Error())
}
