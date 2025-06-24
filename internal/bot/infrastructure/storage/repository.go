package storage

import (
	"log/slog"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

type Repo interface {
	AddLink(link domain.Link)
	DeleteLink(url string, userID int64) bool
	GetLinks(userID int64) []domain.Link
}

type Repository struct {
	links map[int64][]domain.Link
	mu    sync.RWMutex
}

func NewRepository() *Repository {
	links := make(map[int64][]domain.Link)

	return &Repository{
		links: links,
	}
}

func (repo *Repository) AddLink(link domain.Link) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	slog.Debug("Adding new link",
		slog.Int64("user_id", link.UserID),
		slog.String("url", link.URL),
		slog.Int("current_links", len(repo.links[link.UserID])),
	)

	repo.links[link.UserID] = append(repo.links[link.UserID], link)

	slog.Info("Link added successfully",
		slog.Int64("user_id", link.UserID),
		slog.String("url", link.URL),
		slog.Int("total_links", len(repo.links[link.UserID])),
	)
}

func (repo *Repository) DeleteLink(url string, userID int64) bool {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	slog.Debug("Attempting to delete link",
		slog.Int64("user_id", userID),
		slog.String("url", url),
		slog.Int("current_links", len(repo.links[userID])),
	)

	links := repo.links[userID]

	for i, link := range links {
		if link.URL == url {
			repo.links[userID] = append(links[:i], links[i+1:]...)

			slog.Info("Link deleted successfully",
				slog.Int64("user_id", userID),
				slog.String("url", url),
				slog.Int("remaining_links", len(repo.links[userID])),
			)

			return true
		}
	}

	slog.Warn("Link not found for deletion",
		slog.Int64("user_id", userID),
		slog.String("url", url),
	)

	return false
}

func (repo *Repository) GetLinks(userID int64) []domain.Link {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	links := repo.links[userID]

	slog.Debug("Retrieving links",
		slog.Int64("user_id", userID),
		slog.Int("links_count", len(links)),
	)

	return links
}
