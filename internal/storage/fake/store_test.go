package fake

import (
	"context"
	"testing"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

func TestStore_ImplementsStoragePort(t *testing.T) {
	s := &Store{}
	if err := s.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
	_, err := s.GetObservation(context.Background(), 1)
	if err != memory.ErrNotFound {
		t.Fatalf("get: %v", err)
	}
}
