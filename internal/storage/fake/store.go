package fake

import (
	"context"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage"
)

// Store is an in-memory test double for storage.Store.
type Store struct {
	PingErr error
}

func (s *Store) Ping(context.Context) error { return s.PingErr }
func (s *Store) Close() error               { return nil }

func (s *Store) SaveObservation(context.Context, memory.CreateInput) (memory.Observation, error) {
	return memory.Observation{}, nil
}

func (s *Store) GetObservation(context.Context, int64) (memory.Observation, error) {
	return memory.Observation{}, memory.ErrNotFound
}

func (s *Store) ListObservations(context.Context, memory.ListFilter) ([]memory.Observation, error) {
	return nil, nil
}

func (s *Store) SearchObservations(context.Context, memory.SearchFilter) ([]memory.Observation, error) {
	return nil, nil
}

func (s *Store) StartSession(context.Context, string, string) (memory.Session, error) {
	return memory.Session{ID: "fake-session", Status: memory.SessionStatusActive}, nil
}

func (s *Store) EndSession(context.Context, memory.EndSessionInput) (memory.Session, error) {
	return memory.Session{Status: memory.SessionStatusEnded}, nil
}

func (s *Store) GetContext(context.Context, memory.ContextFilter) (memory.ContextResult, error) {
	return memory.ContextResult{}, nil
}

// Compile-time check that Store implements storage.Store.
var _ storage.Store = (*Store)(nil)
