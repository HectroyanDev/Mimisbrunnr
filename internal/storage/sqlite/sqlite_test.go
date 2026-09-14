package sqlite

import (
	"context"
	"testing"
)

func TestOpen_PingCloseReopen(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/mimisbrunnr.db"

	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	store, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping after reopen: %v", err)
	}
}
