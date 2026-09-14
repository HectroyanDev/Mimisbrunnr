package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schemaVersion = 4

// Store is the SQLite implementation of storage.Store.
type Store struct {
	db *sql.DB
}

// Open opens or creates the database at path, runs migrations, and returns a Store.
func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &Store{db: db}
	if err := store.configure(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) configure(ctx context.Context) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := s.db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("sqlite pragma: %w", err)
		}
	}
	return nil
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_meta (
	version INTEGER NOT NULL
);`); err != nil {
		return fmt.Errorf("sqlite migrate meta table: %w", err)
	}

	version, err := s.currentSchemaVersion(ctx)
	if err != nil {
		return err
	}

	if version < 1 {
		if err := s.migrateV1(ctx); err != nil {
			return err
		}
		version = 1
	}
	if version < 2 {
		if err := s.migrateV2(ctx); err != nil {
			return err
		}
		version = 2
	}
	if version < 3 {
		if err := s.migrateV3(ctx); err != nil {
			return err
		}
		version = 3
	}
	if version < 4 {
		if err := s.migrateV4(ctx); err != nil {
			return err
		}
		version = 4
	}
	if version > schemaVersion {
		return fmt.Errorf("sqlite migrate: unsupported schema version %d", version)
	}
	if version < schemaVersion {
		if err := s.setSchemaVersion(ctx, schemaVersion); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) migrateV1(ctx context.Context) error {
	return s.setSchemaVersion(ctx, 1)
}

func (s *Store) migrateV2(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS observations (
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_observations_topic
	ON observations (project, scope, topic_key)
	WHERE topic_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_observations_project_updated
	ON observations (project, updated_at DESC);
`
	if _, err := s.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("sqlite migrate v2: %w", err)
	}
	return s.setSchemaVersion(ctx, 2)
}

func (s *Store) migrateV3(ctx context.Context) error {
	const ddl = `
CREATE VIRTUAL TABLE IF NOT EXISTS observations_fts USING fts5(
	title,
	content
);

INSERT OR REPLACE INTO observations_fts(rowid, title, content)
SELECT id, title, content FROM observations;
`
	if _, err := s.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("sqlite migrate v3: %w", err)
	}
	return s.setSchemaVersion(ctx, 3)
}

func (s *Store) migrateV4(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	project TEXT NOT NULL,
	scope TEXT NOT NULL DEFAULT 'project',
	status TEXT NOT NULL DEFAULT 'active',
	started_at TEXT NOT NULL,
	ended_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_project_started
	ON sessions (project, started_at DESC);

ALTER TABLE observations ADD COLUMN session_id TEXT;

CREATE INDEX IF NOT EXISTS idx_observations_session_updated
	ON observations (session_id, updated_at DESC);
`
	if _, err := s.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("sqlite migrate v4: %w", err)
	}
	return s.setSchemaVersion(ctx, 4)
}

func (s *Store) currentSchemaVersion(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_meta").Scan(&count); err != nil {
		return 0, fmt.Errorf("sqlite schema version count: %w", err)
	}
	if count == 0 {
		return 0, nil
	}

	var version int
	if err := s.db.QueryRowContext(ctx, "SELECT version FROM schema_meta LIMIT 1").Scan(&version); err != nil {
		return 0, fmt.Errorf("sqlite schema version read: %w", err)
	}
	return version, nil
}

func (s *Store) setSchemaVersion(ctx context.Context, version int) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM schema_meta"); err != nil {
		return fmt.Errorf("sqlite schema version reset: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, "INSERT INTO schema_meta (version) VALUES (?)", version); err != nil {
		return fmt.Errorf("sqlite schema version set: %w", err)
	}
	return nil
}

// Ping verifies the database connection and schema bootstrap.
func (s *Store) Ping(ctx context.Context) error {
	var version int
	if err := s.db.QueryRowContext(ctx, "SELECT version FROM schema_meta LIMIT 1").Scan(&version); err != nil {
		return fmt.Errorf("sqlite ping: %w", err)
	}
	if version != schemaVersion {
		return fmt.Errorf("sqlite ping: unexpected schema version %d", version)
	}
	return nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}
