package repository

import (
	"context"
	"time"
)

type (
	OutboxMessage struct {
		ID        uint
		EventType string
		Payload   []byte
		CreatedAt time.Time
	}
)

type OutboxRepository interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkAsSent(ctx context.Context, id int) error
}
