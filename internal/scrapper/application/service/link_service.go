package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/scheduler"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type (
	Link struct {
		URL string
	}
)

type LinkService interface {
	GetAllLinks(ctx context.Context, userID int) ([]Link, error)
	AddLink(ctx context.Context, userID int, link string) (int, error)
	DeleteLink(ctx context.Context, userID int, link string) error
}

type linkService struct {
	linkRepo  repository.LinkRepository
	txManager repository.TxManager
	scheduler scheduler.Scheduler
	logger    *slog.Logger
}

func NewLinkService(linkRepo repository.LinkRepository,
	txManager repository.TxManager, logger *slog.Logger, sched scheduler.Scheduler) LinkService {
	return &linkService{
		linkRepo:  linkRepo,
		scheduler: sched,
		txManager: txManager,
		logger:    logger,
	}
}

func (s *linkService) AddLink(ctx context.Context, userID int, link string) (int, error) {
	s.logger.Debug("Starting to add link",
		slog.Int("user_id", userID),
		slog.String("link", link),
		slog.String("operation", "add_link"),
	)

	var linkID int

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		links, err := s.linkRepo.GetAllLinks(txCtx, userID)
		if err != nil {
			s.logger.Error("Failed to get existing links",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error receiving links: %s", err.Error())
		}

		s.logger.Debug("Retrieved existing links",
			slog.Int("user_id", userID),
			slog.Int("existing_links_count", len(links)),
		)

		for _, l := range links {
			if l == link {
				return fmt.Errorf("Link already exists")
			}
		}

		linkID, err = s.linkRepo.AddLink(txCtx, userID, link)
		if err != nil {
			s.logger.Error("Failed to add link to repository",
				slog.Int("user_id", userID),
				slog.String("link", link),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("link addition error: %s", err.Error())
		}

		if err := s.scheduler.AddLink(link, userID, linkID); err != nil {
			s.logger.Error("Failed to add link to scheduler",
				slog.Int("user_id", userID),
				slog.Int("link_id", linkID),
				slog.String("link", link),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error adding a link to the scheduler: %s", err.Error())
		}

		s.logger.Info("Link added successfully",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("link", link),
		)

		return nil
	})

	if err != nil {
		return -1, err
	}

	return linkID, nil
}

func (s *linkService) DeleteLink(ctx context.Context, userID int, link string) error {
	s.logger.Debug("Starting to delete link",
		slog.Int("user_id", userID),
		slog.String("link", link),
		slog.String("operation", "delete_link"),
	)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		linkID, err := s.linkRepo.GetLinkIDByURL(txCtx, userID, link)
		if err != nil {
			s.logger.Error("Failed to find link by URL",
				slog.Int("user_id", userID),
				slog.String("link", link),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("the link was not found")
		}

		err = s.linkRepo.DeleteLink(txCtx, userID, linkID, link)
		if err != nil {
			s.logger.Error("Failed to delete link from repository",
				slog.Int("user_id", userID),
				slog.Int("link_id", linkID),
				slog.String("link", link),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error when deleting a link: %s", err.Error())
		}

		err = s.scheduler.RemoveLink(link)
		if err != nil {
			s.logger.Error("Failed to remove link from scheduler",
				slog.Int("user_id", userID),
				slog.Int("link_id", linkID),
				slog.String("link", link),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error when deleting a link from scheduler: %s", err.Error())
		}

		s.logger.Info("Link deleted successfully",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("link", link),
		)

		return nil
	})
}

func (s *linkService) GetAllLinks(ctx context.Context, userID int) ([]Link, error) {
	s.logger.Debug("Getting all links for user",
		slog.Int("user_id", userID),
		slog.String("operation", "get_all_links"),
	)

	var rawLinks []string

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		var err error
		rawLinks, err = s.linkRepo.GetAllLinks(txCtx, userID)

		if err != nil {
			s.logger.Error("Failed to get links from repository",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("error receiving links: %w", err)
		}

		s.logger.Debug("Retrieved links from repository",
			slog.Int("user_id", userID),
			slog.Int("links_count", len(rawLinks)),
		)

		return nil
	})

	if err != nil {
		return nil, err
	}

	links := make([]Link, 0, len(rawLinks))
	for _, url := range rawLinks {
		links = append(links, Link{URL: url})
	}

	s.logger.Info("Successfully retrieved user links",
		slog.Int("user_id", userID),
		slog.Int("links_count", len(links)),
	)

	return links, nil
}
