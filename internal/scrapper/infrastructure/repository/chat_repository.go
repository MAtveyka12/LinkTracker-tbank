package repository

import "context"

type ChatRepository interface {
	RegisterChat(ctx context.Context, userID int) error
	DeleteChat(ctx context.Context, userID int) error
	ChatExists(ctx context.Context, userID int) (bool, error)
}
