package scrapper

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

type ScraperClient interface {
	GetLinks(ctx context.Context, userID int64) ([]string, error)
	RegisterUser(ctx context.Context, user domain.User) error
	AddLink(ctx context.Context, link domain.Link) error
	RemoveLink(ctx context.Context, link string, userID int64) error
}
