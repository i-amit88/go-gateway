package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrSessionNotFound = errors.New("session not found")

type MemorySessionManager struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessions: make(map[string]Session),
	}
}

func (m *MemorySessionManager) Create(ctx context.Context, workingDir string) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ValidateWorkingDir(workingDir); err != nil {
		return nil, err
	}

	now := time.Now()
	session := Session{
		ID:         uuid.NewString(),
		Status:     SessionCreated,
		WorkingDir: workingDir,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return &session, nil
}

func (m *MemorySessionManager) Get(ctx context.Context, id string) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	session, exists := m.sessions[id]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

func (m *MemorySessionManager) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; !exists {
		return ErrSessionNotFound
	}

	delete(m.sessions, id)

	return nil
}
