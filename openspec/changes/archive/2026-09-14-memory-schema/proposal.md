# Proposal: memory-schema

## Why

Mimisbrunnr already exposes MCP stdio and a SQLite storage port with only a health bootstrap (`schema_meta`). To persist agent memory, we need a domain model and database schema for observations before exposing `mem_*` MCP tools. This change lays that foundation so later work can add tools and search without redesigning persistence.

## What Changes

- Introduce domain type `Observation` and validation rules in Go.
- Add SQLite migration v2: `observations` table with indexes.
- Extend `storage.Store` with create/read/list methods for observations (no MCP tools yet).
- Bump schema version from 1 to 2; existing databases migrate forward on open.
- **Non-goals for this change**: MCP tools (`mem_save`, `mem_search`), FTS5/full-text search, sessions, relations/conflicts, cloud sync.

## Capabilities

### New Capabilities

- `memory-observations`: Domain model, validation, and storage contract for persisting and retrieving observations locally.

### Modified Capabilities

- `storage-port`: Extend `Store` beyond `Ping`/`Close`; add migration v2; replace REQ-206 (no memory schema) with observation persistence requirements.

## Impact

- **Code**: `internal/memory` (new), `internal/storage/store.go`, `internal/storage/sqlite/`, tests.
- **API**: `storage.Store` interface grows; breaking for any external implementers (none today).
- **Database**: SQLite file gains `observations` table; `schema_meta.version` becomes `2`.
- **MCP**: unchanged in this change (`ping` only).
- **Dependencies**: none new (continues `modernc.org/sqlite`).
