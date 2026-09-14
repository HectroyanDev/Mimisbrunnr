## 1. Domain (`internal/memory`)

- [x] 1.1 Add `Observation`, `CreateInput`, `ListFilter`, and validation helpers with table-driven tests for required fields, scope enum, and max lengths — verify `go test ./internal/memory/...` passes
- [x] 1.2 Add `ErrNotFound` and validation error types — verify tests assert error kinds for invalid input and missing IDs

## 2. Storage port

- [x] 2.1 Extend `storage.Store` with `SaveObservation`, `GetObservation`, `ListObservations` — verify `go build ./...` succeeds
- [x] 2.2 Add fake/in-memory store test double implementing the extended interface for MCP tests — verify compile and use in at least one test

## 3. SQLite migration v2

- [x] 3.1 Bump `schemaVersion` to 2 and implement idempotent migration v2 (`observations` table + indexes) — verify `internal/storage/sqlite` test upgrades from v1 TempDir DB
- [x] 3.2 Update `Ping` to require version 2 — verify ping fails on unsupported version and passes after migration

## 4. SQLite observation persistence

- [x] 4.1 Implement `SaveObservation` insert path — verify test creates row with timestamps and assigned ID
- [x] 4.2 Implement topic-key upsert (update existing row) — verify second save with same topic_key updates content and `updated_at` but preserves `id`/`created_at`
- [x] 4.3 Implement `GetObservation` and `ListObservations` (project filter, optional scope, limit, newest first) — verify list ordering and not-found behavior

## 5. Integration and docs

- [x] 5.1 Ensure existing `storage.Open` and MCP `ping` tests still pass — verify `go test ./...`
- [x] 5.2 Update `AGENTS.md` baseline capabilities list after archive (defer until `/opsx-archive`) — verify N/A until archive; skip if applying only

## 6. Verification gate

- [x] 6.1 Run full test suite and build — verify `go test ./...` and `go build ./cmd/mimisbrunnr` succeed
- [x] 6.2 Run `openspec validate --change memory-schema` — verify change validates with zero failures
