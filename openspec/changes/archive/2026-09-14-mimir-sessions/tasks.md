# Tasks: mimir-sessions

## 1. Domain

- [x] 1.1 Add session types, context filter/entry, snippet helper, validation — verify `go test ./internal/memory/...`
- [x] 1.2 Extend `CreateInput` with `SessionID` validation — verify tests

## 2. Storage

- [x] 2.1 Migration v4 (`sessions` table, `observations.session_id`) — verify migrate tests
- [x] 2.2 Implement `StartSession`, `EndSession`, `GetContext` in sqlite — verify sqlite tests
- [x] 2.3 Link `session_id` in `SaveObservation` — verify upsert + session tests
- [x] 2.4 Extend `storage.Store` and fake store — verify compile

## 3. MCP

- [x] 3.1 Register `mimir_session_start`, `mimir_session_end`, `mimir_context` — verify MCP tests
- [x] 3.2 Extend `mimir_save` with `session_id` — verify round-trip test
- [x] 3.3 Update server tool list test

## 4. Docs & validation

- [x] 4.1 Update `AGENTS.md`, `README.md`, `openspec/config.yaml`
- [x] 4.2 `openspec validate mimir-sessions --strict` and `go test ./...`
