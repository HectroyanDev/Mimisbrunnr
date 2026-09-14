# Design: mimir-sessions

## Context

Baseline: observations + FTS (schema v3), MCP tools `mimir_save/get/list/search`. Inspired by OpenViking session commit + L0/L1 loading, without LLM extraction or `viking://` filesystem.

## Schema v4

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    project TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'project',
    status TEXT NOT NULL DEFAULT 'active',
    started_at TEXT NOT NULL,
    ended_at TEXT
);

CREATE INDEX idx_sessions_project_started ON sessions (project, started_at DESC);

ALTER TABLE observations ADD COLUMN session_id TEXT REFERENCES sessions(id);
CREATE INDEX idx_observations_session_updated ON observations (session_id, updated_at DESC);
```

SQLite `ALTER ADD COLUMN` only; no FK enforcement at DB level beyond declaration.

## Domain (`internal/memory`)

- `Session`, `SessionStatusActive`, `SessionStatusEnded`
- `ContextFilter` — `Project`, `SessionID`, `Limit`
- `ContextEntry` — `ID`, `Title`, `Snippet`, `Type`, `TopicKey`, `UpdatedAt` (no full content)
- `EndSessionInput` — `SessionID`, `Summary`
- Extend `CreateInput` with optional `SessionID`
- Constants: `DefaultContextLimit=10`, `MaxContextLimit=50`, `SnippetLen=500`, `TypeSessionSummary="session_summary"`

## Store interface

```go
StartSession(ctx, project, scope) (Session, error)
EndSession(ctx, EndSessionInput) (Session, error)
GetContext(ctx, ContextFilter) (ContextResult, error)
```

`ContextResult` holds `ContextText` (formatted markdown-ish block) and `Entries []ContextEntry`.

## Session lifecycle

1. `StartSession` → insert `sessions` row, UUID v4.
2. `SaveObservation` with `session_id` → validate session active + project match; set column; bump nothing else.
3. `EndSession` → set ended; `SaveObservation` with `type=session_summary`, `topic_key=sessions/{id}/summary`, same project/scope, `session_id` set.
4. `GetContext` → query observations (session filter if set), build snippets with `truncateSnippet(content, 500)`.

## MCP (`internal/mcp`)

| Tool | Handler file |
|------|----------------|
| `mimir_session_start` | `tools_session.go` |
| `mimir_session_end` | `tools_session.go` |
| `mimir_context` | `tools_context.go` |

Extend `mimirSaveInput` with `session_id`.

## `context_text` format

```
# Mímir's well — recent context

## [title] (type)
snippet...

---
```

## Topic key convention (documented, not enforced in code)

- Session summary: `sessions/{session_id}/summary`
- Themed paths: `{area}/{name}` e.g. `architecture/auth-model`

## Risks

| Risk | Mitigation |
|------|------------|
| Orphan `session_id` on observations after manual DB edits | Validate on save only |
| `mimir_context` vs `mimir_list` overlap | Context omits full content; lower default limit |
| Migration v4 on large DBs | `ALTER ADD COLUMN` is O(1) metadata in SQLite |

## Non-goals (unchanged)

LLM extraction, message archive, peers, memory diff audit table.
