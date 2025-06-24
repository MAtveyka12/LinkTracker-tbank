package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/mock"
)

const (
	link = "https://example.com"
)

func TestAddLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().AddLink(ctx, userID, link).Return(1, nil)
	id, err := mockRepo.AddLink(ctx, userID, link)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestDeleteLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1
	linkID := 1

	mockRepo.EXPECT().DeleteLink(ctx, userID, linkID, link).Return(nil)
	err := mockRepo.DeleteLink(ctx, userID, linkID, link)
	assert.NoError(t, err)
}

func TestDeleteLink_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1
	linkID := 2

	mockRepo.EXPECT().DeleteLink(ctx, userID, linkID, link).Return(errors.New("ссылка не найдена"))
	err := mockRepo.DeleteLink(ctx, userID, linkID, link)
	assert.Error(t, err)
}

func TestGetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1
	linkID := 1

	mockRepo.EXPECT().GetLink(ctx, userID, linkID).Return(link, linkID, nil)

	url, id, err := mockRepo.GetLink(ctx, userID, linkID)
	assert.NoError(t, err)
	assert.Equal(t, link, url)
	assert.Equal(t, linkID, id)
}

func TestGetAllLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1
	links := []string{"https://example1.com", "https://example2.com"}

	mockRepo.EXPECT().GetAllLinks(ctx, userID).Return(links, nil)

	result, err := mockRepo.GetAllLinks(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, links, result)
}

func TestRegisterChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()
	userID := 1

	mockChatRepo.EXPECT().RegisterChat(ctx, userID).Return(nil)

	err := mockChatRepo.RegisterChat(ctx, userID)
	assert.NoError(t, err)
}
