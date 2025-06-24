package inmemory

import (
	"context"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

type StateManagerImp struct {
	trackStates   map[int64]string
	untrackStates map[int64]bool
	mu            sync.RWMutex
}

func NewStateManager() domain.StateManager {
	return &StateManagerImp{
		trackStates:   make(map[int64]string),
		untrackStates: make(map[int64]bool),
	}
}

func (sm *StateManagerImp) SetTrackState(_ context.Context, userID int64, state string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.trackStates[userID] = state

	return nil
}

func (sm *StateManagerImp) GetTrackState(_ context.Context, userID int64) (state string, exists bool, err error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	state, exists = sm.trackStates[userID]

	return state, exists, nil
}

func (sm *StateManagerImp) DeleteTrackState(_ context.Context, userID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.trackStates, userID)

	return nil
}

func (sm *StateManagerImp) SetUntrackState(_ context.Context, userID int64, state bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.untrackStates[userID] = state

	return nil
}

func (sm *StateManagerImp) GetUntrackState(_ context.Context, userID int64) (state, exists bool, err error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	state, exists = sm.untrackStates[userID]

	return state, exists, nil
}

func (sm *StateManagerImp) DeleteUntrackState(_ context.Context, userID int64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.untrackStates, userID)

	return nil
}
