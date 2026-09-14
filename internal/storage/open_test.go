package storage

import (
	"context"
	"strings"
	"testing"
)

func TestOpen_UnsupportedDriver(t *testing.T) {
	_, err := Open(context.Background(), Config{Driver: "postgres"})
	if err == nil {
		t.Fatal("expected error for postgres driver")
	}
	if !strings.Contains(err.Error(), "unsupported storage driver") {
		t.Fatalf("error: %v", err)
	}
}

func TestOpen_DefaultsToSQLite(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/test.db"

	store, err := Open(ctx, Config{SQLitePath: path})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}
