package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"

	_ "modernc.org/sqlite"
)

func TestMigrate_UpgradeFromV1(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/v1.db"

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open v1 db: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE schema_meta (version INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create schema_meta: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_meta(version) VALUES (1)`); err != nil {
		t.Fatalf("seed v1: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close v1 db: %v", err)
	}

	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open upgraded: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	var name string
	if err := store.db.QueryRowContext(ctx, `
SELECT name FROM sqlite_master WHERE type='table' AND name='observations'`).Scan(&name); err != nil {
		t.Fatalf("observations table missing: %v", err)
	}
	if err := store.db.QueryRowContext(ctx, `
SELECT name FROM sqlite_master WHERE type='table' AND name='observations_fts'`).Scan(&name); err != nil {
		t.Fatalf("observations_fts table missing: %v", err)
	}
}

func TestMigrate_UpgradeFromV2BackfillsFTS(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/v2.db"

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open v2 db: %v", err)
	}
	const v2DDL = `
CREATE TABLE schema_meta (version INTEGER NOT NULL);
INSERT INTO schema_meta(version) VALUES (2);
CREATE TABLE observations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	type TEXT NOT NULL,
	project TEXT NOT NULL,
	scope TEXT NOT NULL DEFAULT 'project',
	topic_key TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
INSERT INTO observations (title, content, type, project, scope, created_at, updated_at)
VALUES ('Legacy', 'alpha token', 'manual', 'mimisbrunnr', 'project', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
`
	if _, err := db.Exec(v2DDL); err != nil {
		t.Fatalf("seed v2: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close v2 db: %v", err)
	}

	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open upgraded: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	rows, err := store.SearchObservations(ctx, memory.SearchFilter{
		Project: "mimisbrunnr",
		Query:   "alpha",
	})
	if err != nil {
		t.Fatalf("search backfill: %v", err)
	}
	if len(rows) != 1 || rows[0].Title != "Legacy" {
		t.Fatalf("backfill search: %#v", rows)
	}
}

func TestPing_UnsupportedSchemaVersion(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/future.db"

	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if _, err := store.db.ExecContext(ctx, "DELETE FROM schema_meta"); err != nil {
		t.Fatalf("reset meta: %v", err)
	}
	if _, err := store.db.ExecContext(ctx, "INSERT INTO schema_meta(version) VALUES (99)"); err != nil {
		t.Fatalf("set future version: %v", err)
	}

	if err := store.Ping(ctx); err == nil {
		t.Fatal("expected ping error for unsupported schema version")
	}
}
