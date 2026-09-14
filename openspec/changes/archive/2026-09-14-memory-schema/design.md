# Design: memory-schema

## Context

See [proposal.md](./proposal.md). Baseline code has `storage.Store` with `Ping`/`Close`, SQLite `schema_meta` at version 1, and MCP `ping` only. This design adds observation persistence behind the same port.

## Goals / Non-Goals

**Goals:**

- Domain package `internal/memory` with `Observation`, input types, and validation.
- SQLite migration v2 (`observations` table, indexes, version bump).
- Extend `storage.Store` with `SaveObservation`, `GetObservation`, `ListObservations`.
- Table-driven tests with `t.TempDir()` for DB paths; strict TDD.

**Non-Goals:**

- MCP tools (`mem_save`, `mem_search`), FTS5, sessions, soft-delete, relations/conflicts.
- Postgres adapter implementation (interface must remain driver-agnostic).
- Changing MCP `ping` behavior beyond existing store ping.

## Decisions

### 1. Package layout

- `internal/memory` — domain types, validation, errors (`ErrNotFound`, `ValidationError`).
- `internal/storage` — extended `Store` interface and input/output DTOs (or aliases to `memory` types).
- `internal/storage/sqlite` — SQL, migrations, upsert logic.

**Alternative considered:** put domain inside `storage`. Rejected: keeps SQL out of domain and matches hexagonal layout from baseline specs.

### 2. Schema version strategy

- Constant `schemaVersion = 2` in sqlite package.
- `migrate()` runs versioned steps: v1 bootstrap (idempotent), v2 `observations`.
- `Ping` checks `schema_meta.version == schemaVersion`.

**Alternative:** single DDL file without version steps. Rejected: harder to add postgres later and to test upgrades from v1.

### 3. Observations DDL (SQLite)

```sql
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
```

Timestamps stored as RFC3339 UTC strings for simplicity (consistent with Engram patterns, easy to test).

### 4. Upsert implementation

When `TopicKey` is non-empty on save:

1. `SELECT id FROM observations WHERE project=? AND scope=? AND topic_key=?`
2. If found → `UPDATE` title, content, type, updated_at.
3. Else → `INSERT`.

Use a transaction per save.

### 5. Store interface extension

```go
SaveObservation(ctx context.Context, in memory.CreateInput) (memory.Observation, error)
GetObservation(ctx context.Context, id int64) (memory.Observation, error)
ListObservations(ctx context.Context, filter memory.ListFilter) ([]memory.Observation, error)
```

`ListFilter`: `Project` (required), `Scope` (optional), `Limit` (default 20, max 100).

### 6. List ordering

`ORDER BY updated_at DESC` with `LIMIT`. No cursor pagination in v2.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking `Store` interface for future adapters | Only sqlite implements for now; compile-time enforcement via tests |
| Topic key upsert race under concurrency | SQLite single-writer acceptable for local MCP; document limitation |
| No FTS → poor recall at scale | Explicit non-goal; next change adds search |
| Config YAML rules format broke OpenSpec parse | Fix `openspec/config.yaml` apply/verify rules to string arrays |

## Migration Plan

1. Ship migration v2 in sqlite `Open()` — automatic on next process start.
2. Existing users with v1 DB: first open runs v2 DDL, version → 2.
3. Rollback: no down-migration; restore DB file from backup if needed (document in README change note at apply time).

## Open Questions

None blocking implementation. FTS and MCP tools are intentionally deferred to follow-on changes.
