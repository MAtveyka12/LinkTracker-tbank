package domain

import (
	"context"
)

const (
	StateWaitingForLink    = "waiting_for_link"
	StateWaitingForTags    = "waiting_for_tags"
	StateWaitingForFilters = "waiting_for_filters"

	StateWaitingForEmail = iota
	StateWaitingForPassword
)

type StateManager interface {
	SetTrackState(ctx context.Context, userID int64, state string) error
	GetTrackState(ctx context.Context, userID int64) (string, bool, error)
	DeleteTrackState(ctx context.Context, userID int64) error

	SetUntrackState(ctx context.Context, userID int64, state bool) error
	GetUntrackState(ctx context.Context, userID int64) (bool, bool, error)
	DeleteUntrackState(ctx context.Context, userID int64) error
}
