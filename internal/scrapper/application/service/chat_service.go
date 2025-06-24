package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type ChatService interface {
	RegisterChat(ctx context.Context, userID int) error
	DeleteChat(ctx context.Context, userID int) error
}

type chatService struct {
	chatRepo  repository.ChatRepository
	linkRepo  repository.LinkRepository
	txManager repository.TxManager
	logger    *slog.Logger
}

func NewChatService(chatRepo repository.ChatRepository, linkRepo repository.LinkRepository,
	txManager repository.TxManager, logger *slog.Logger) ChatService {
	return &chatService{
		chatRepo:  chatRepo,
		linkRepo:  linkRepo,
		txManager: txManager,
		logger:    logger,
	}
}

func (s *chatService) RegisterChat(ctx context.Context, userID int) error {
	s.logger.Debug("Starting chat registration",
		slog.Int("user_id", userID),
		slog.String("operation", "register_chat"),
	)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		links, err := s.linkRepo.GetAllLinks(txCtx, userID)
		if err == nil && len(links) > 0 {
			s.logger.Warn("Chat already registered with existing links",
				slog.Int("user_id", userID),
				slog.Int("links_count", len(links)),
			)

			return fmt.Errorf("the chat has already been registered")
		}

		err = s.chatRepo.RegisterChat(txCtx, userID)
		if err != nil {
			s.logger.Error("Failed to register chat",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("chat creation error: %s", err.Error())
		}

		s.logger.Info("Chat registered successfully",
			slog.Int("user_id", userID),
		)

		return nil
	})
}

func (s *chatService) DeleteChat(ctx context.Context, userID int) error {
	s.logger.Debug("Starting chat deletion",
		slog.Int("user_id", userID),
		slog.String("operation", "delete_chat"),
	)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		_, err := s.linkRepo.GetAllLinks(txCtx, userID)
		if err != nil {
			s.logger.Error("Failed to get links for chat",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error receiving links: %s", err.Error())
		}

		s.logger.Debug("Found links to delete",
			slog.Int("user_id", userID),
		)

		err = s.linkRepo.DeleteAllLinks(txCtx, userID)
		if err != nil {
			s.logger.Error("Failed to delete links",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("link deletion error: %s", err.Error())
		}

		s.logger.Info("Chat deleted successfully",
			slog.Int("user_id", userID),
		)

		return nil
	})
}
