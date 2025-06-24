package repository

import "context"

type LinkRepository interface {
	GetLink(ctx context.Context, userID, linkID int) (string, int, error)
	GetAllLinks(ctx context.Context, userID int) ([]string, error)
	AddLink(ctx context.Context, userID int, link string) (int, error)
	DeleteLink(ctx context.Context, userID, linkID int, link string) error
	DeleteAllLinks(ctx context.Context, userID int) error
	GetLinkIDByURL(ctx context.Context, userID int, url string) (int, error)
}
