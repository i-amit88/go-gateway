package session

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidWorkingDir = errors.New("invalid working directory")

type SessionStatus string

const (
	SessionCreated SessionStatus = "created"
	SessionActive  SessionStatus = "active"
	SessionStopped SessionStatus = "stopped"
)

type Session struct {
	ID         string
	Status     SessionStatus
	WorkingDir string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SessionManager interface {
	Create(ctx context.Context, workingDir string) (*Session, error)
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
}
