# Proposal: mimir-sessions

## Why

Agents lose thread after compaction or between turns. Mimisbrunnr already persists observations and supports FTS, but there is no session boundary or a cheap way to reload recent work without pulling full `content` for every row. OpenViking-inspired tiered context (abstract + snippet before full detail) fits the `mimir_*` theme without adopting its full filesystem or LLM extraction pipeline.

## What Changes

- Add lightweight **sessions** (start, optional link on save, end with summary).
- Add **`mimir_context`** returning L0/L1-style entries (title + snippet) plus readable `context_text`.
- Add **`mimir_session_start`** and **`mimir_session_end`** MCP tools.
- Extend **`mimir_save`** with optional `session_id` (must reference an active session).
- SQLite migration **v4**: `sessions` table + nullable `session_id` on `observations`.
- Extend `storage.Store` with session and context methods.

## Non-goals

- LLM-based memory extraction or async background jobs (OpenViking `commit` phase 2).
- Virtual filesystem (`mimir://` URIs) — only document `topic_key` conventions.
- Message log / `add_message` API.
- Changes to `mimir_search` ranking or FTS schema.

## Capabilities

### Modified Capabilities

- `memory-observations`: Session domain, context retrieval tiers, `session_summary` type convention.
- `storage-port`: Migration v4, session persistence APIs.
- `mcp-server`: Register `mimir_session_start`, `mimir_session_end`, `mimir_context`; extend `mimir_save`.

## Impact

- **Database**: `schema_meta.version` becomes `4`.
- **API**: `storage.Store` and `memory.CreateInput` grow; fake store updated.
- **MCP**: Three new tools; `mimir_save` gains optional `session_id`.
