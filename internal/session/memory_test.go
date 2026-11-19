package session

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
)

func TestMemorySessionManagerCreatesAndGetsSession(t *testing.T) {
	manager := NewMemorySessionManager()
	workingDir := t.TempDir()

	created, err := manager.Create(context.Background(), workingDir)
	if err != nil {
		t.Fatalf("create returned unexpected error: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected session ID")
	}
	if created.Status != SessionCreated {
		t.Fatalf("expected status %q, got %q", SessionCreated, created.Status)
	}

	got, err := manager.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get returned unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, got.ID)
	}
	if got.WorkingDir != workingDir {
		t.Fatalf("expected working directory %q, got %q", workingDir, got.WorkingDir)
	}
}

func TestMemorySessionManagerCreatesUniqueIDs(t *testing.T) {
	manager := NewMemorySessionManager()
	workingDir := t.TempDir()
	ids := make(map[string]struct{})

	for i := 0; i < 100; i++ {
		created, err := manager.Create(context.Background(), workingDir)
		if err != nil {
			t.Fatalf("create returned unexpected error: %v", err)
		}
		if _, exists := ids[created.ID]; exists {
			t.Fatalf("duplicate session ID %q", created.ID)
		}
		ids[created.ID] = struct{}{}
	}
}

func TestMemorySessionManagerDeletesSession(t *testing.T) {
	manager := NewMemorySessionManager()
	created, err := manager.Create(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("create returned unexpected error: %v", err)
	}

	if err := manager.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete returned unexpected error: %v", err)
	}

	_, err = manager.Get(context.Background(), created.ID)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}

	err = manager.Delete(context.Background(), created.ID)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound on second delete, got %v", err)
	}
}

func TestMemorySessionManagerReturnsContextCancellation(t *testing.T) {
	manager := NewMemorySessionManager()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := manager.Create(ctx, t.TempDir()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled from create, got %v", err)
	}
	if _, err := manager.Get(ctx, "missing"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled from get, got %v", err)
	}
	if err := manager.Delete(ctx, "missing"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled from delete, got %v", err)
	}
}

func TestMemorySessionManagerSupportsConcurrentCreates(t *testing.T) {
	manager := NewMemorySessionManager()
	workingDir := t.TempDir()
	const workers = 50

	var waitGroup sync.WaitGroup
	errorsFound := make(chan error, workers)
	waitGroup.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer waitGroup.Done()
			if _, err := manager.Create(context.Background(), workingDir); err != nil {
				errorsFound <- err
			}
		}()
	}

	waitGroup.Wait()
	close(errorsFound)

	for err := range errorsFound {
		t.Errorf("concurrent create returned error: %v", err)
	}
}

func TestValidateWorkingDirAcceptsExistingDirectory(t *testing.T) {
	if err := ValidateWorkingDir(t.TempDir()); err != nil {
		t.Fatalf("expected existing directory to be valid, got %v", err)
	}
}

func TestValidateWorkingDirRejectsInvalidPaths(t *testing.T) {
	filePath := t.TempDir() + "/file.txt"
	if err := os.WriteFile(filePath, []byte("data"), 0600); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "empty", path: ""},
		{name: "relative", path: "./project"},
		{name: "missing", path: t.TempDir() + "/missing"},
		{name: "file", path: filePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateWorkingDir(tt.path); !errors.Is(err, ErrInvalidWorkingDir) {
				t.Fatalf("expected ErrInvalidWorkingDir, got %v", err)
			}
		})
	}
}
