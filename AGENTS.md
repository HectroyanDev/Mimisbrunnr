# Mimisbrunnr — Agent Guide

Mimisbrunnr is a Go MCP server for local agent memory at Mímir's well. Agents interact through **`mimir_*` tools** — not `mem_*` (Engram-style). This document is an index; behavioral requirements live in OpenSpec.

## Source of truth

| What | Where |
|------|-------|
| Current product behavior | [`openspec/specs/`](openspec/specs/) |
| Planned or in-flight work | [`openspec/changes/<name>/`](openspec/changes/) (create with `openspec new change <name>`) |
| Project context and SDD rules | [`openspec/config.yaml`](openspec/config.yaml) |

Do not implement features that contradict baseline specs. New behavior starts as a change (proposal → specs delta → design → tasks) before apply.

## Mimir tools (agent API)

All MCP tools use the `mimir_` prefix. The agent is the actor at the well.

| Tool | Status | Maps to |
|------|--------|---------|
| `mimir_ping` | implemented | `storage.Store.Ping` |
| `mimir_save` | implemented | `SaveObservation` |
| `mimir_get` | implemented | `GetObservation` |
| `mimir_list` | implemented | `ListObservations` |
| `mimir_search` | implemented | `SearchObservations` (FTS5) |
| `mimir_session_start` | implemented | `StartSession` |
| `mimir_session_end` | implemented | `EndSession` |
| `mimir_context` | implemented | `GetContext` (tiered) |

Domain types stay in `internal/memory` (`Observation`, etc.); only the MCP surface is themed.

## Baseline capabilities

- [`openspec/specs/mcp-server/spec.md`](openspec/specs/mcp-server/spec.md) — stdio MCP, `mimir_*` naming, `mimir_ping`
- [`openspec/specs/storage-port/spec.md`](openspec/specs/storage-port/spec.md) — `Store` port, SQLite v4, env config
- [`openspec/specs/memory-observations/spec.md`](openspec/specs/memory-observations/spec.md) — observation domain, save/get/list, topic-key upsert

## Layout

```
cmd/mimisbrunnr/           MCP binary (stdio)
internal/memory/           Observation domain and validation
internal/mcp/              MCP tools and server wiring
internal/storage/          Store interface and factory
internal/storage/sqlite/   SQLite adapter
internal/config/           Environment configuration
openspec/                  Spec-driven development artifacts
```

## Commands

```bash
go test ./...
go build ./cmd/mimisbrunnr
go run ./cmd/mimisbrunnr
openspec validate
openspec list --specs
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `MIMISBRUNNR_DRIVER` | `sqlite` | Storage driver (only `sqlite` implemented) |
| `MIMISBRUNNR_SQLITE_PATH` | `~/.config/mimisbrunnr/mimisbrunnr.db` | SQLite file path |

## OpenSpec skills (Cursor)

- `.cursor/skills/openspec-explore` — explore before proposing
- `.cursor/skills/openspec-propose` — new change with artifacts
- `.cursor/skills/openspec-apply-change` — implement from tasks
- `.cursor/skills/openspec-archive-change` — merge deltas into specs

## Conventions

- MCP tools: `mimir_<verb>` only; never `mem_*`.
- Domain code depends on `storage.Store`, not SQL.
- New storage backends register only through `storage.Open`.
- Tests: table-driven where practical; `t.TempDir()` for filesystem/DB paths.
- Strict TDD: red → green → refactor; `go test ./...` is the gate.
