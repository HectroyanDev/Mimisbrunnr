# Storage Port Specification

## Purpose

Provide a persistence port (`storage.Store`) that hides SQL and driver details from MCP and domain code. The SQLite adapter implements local observation and session persistence via migration v4 (observations + FTS5 + sessions), environment-based configuration, and sequential schema migrations. Alternate drivers (e.g. postgres) remain out of scope.

---

## Requirements

### Requirement: REQ-200 Store Interface

The system MUST expose a `storage.Store` interface with `Ping(ctx context.Context) error` and `Close() error`.

#### Scenario: Store lifecycle

- GIVEN a successfully opened store
- WHEN `Ping` is called with a valid context
- THEN no error is returned AND the store confirms persistence is reachable
- WHEN `Close` is called
- THEN resources are released without panic

---

### Requirement: REQ-201 Storage Factory and Default Driver

`storage.Open` MUST select the storage adapter from `storage.Config.Driver`. An empty driver MUST default to `sqlite`. An unsupported driver MUST return an explicit error.

#### Scenario: Default driver

- GIVEN `Config` with empty `Driver` and a valid `SQLitePath`
- WHEN `Open` is called
- THEN a SQLite-backed store is returned without error

#### Scenario: Unsupported driver

- GIVEN `Config` with `Driver` set to `postgres` (or any value other than `sqlite`)
- WHEN `Open` is called
- THEN an error is returned AND the error message indicates an unsupported driver AND no panic occurs

---

### Requirement: REQ-202 SQLite Adapter Bootstrap

The SQLite adapter MUST use `modernc.org/sqlite` via `database/sql`. On open it MUST create the parent directory of the database file, apply `PRAGMA foreign_keys = ON`, and apply `PRAGMA busy_timeout = 5000`.

#### Scenario: First open on a new path

- GIVEN a database path whose parent directory does not exist
- WHEN the SQLite adapter opens that path
- THEN the parent directory is created AND the database file is usable

---

### Requirement: REQ-203 Schema Migration v1

On open, the SQLite adapter MUST ensure table `schema_meta(version INTEGER NOT NULL)` exists. The adapter MUST apply sequential migrations until the database reaches the current schema version. `Ping` MUST verify that `schema_meta.version` equals the current version (`4` after sessions).

#### Scenario: Fresh database

- GIVEN a new SQLite file
- WHEN the adapter opens and `Ping` is called
- THEN `schema_meta` contains version `4` AND the `sessions` table exists AND `observations.session_id` exists

#### Scenario: Upgrade from version 1

- GIVEN a database at `schema_meta.version = 1` with only bootstrap tables
- WHEN the adapter opens
- THEN migrations run through v4 AND `schema_meta.version` becomes `4` AND existing v1 data is preserved

#### Scenario: Reopen existing database

- GIVEN a database already migrated to version `4`
- WHEN the adapter is closed and reopened on the same path
- THEN `Ping` still succeeds without reinitializing to a different version

#### Scenario: Upgrade from version 2

- GIVEN a database at `schema_meta.version = 2` with existing observations
- WHEN the adapter opens
- THEN migrations v3 and v4 run AND `schema_meta.version` becomes `4`

#### Scenario: Upgrade from version 3

- GIVEN a database at `schema_meta.version = 3`
- WHEN the adapter opens
- THEN migration v4 runs AND `schema_meta.version` becomes `4`

#### Scenario: Unexpected schema version

- GIVEN `schema_meta` contains a version greater than the binary supports
- WHEN `Ping` is called
- THEN an error is returned indicating unsupported schema version

---

### Requirement: REQ-204 Environment Configuration

Runtime configuration MUST be loadable from environment variables:

- `MIMISBRUNNR_DRIVER` — default `sqlite` when unset
- `MIMISBRUNNR_SQLITE_PATH` — default `{UserConfigDir}/mimisbrunnr/mimisbrunnr.db` when unset; if `UserConfigDir` is unavailable, default MUST be `mimisbrunnr.db` in the current working context

#### Scenario: Defaults

- GIVEN neither environment variable is set AND `UserConfigDir` is available
- WHEN configuration is loaded
- THEN `Driver` is `sqlite` AND `SQLitePath` ends with `mimisbrunnr/mimisbrunnr.db` under the user config directory

#### Scenario: Explicit overrides

- GIVEN `MIMISBRUNNR_DRIVER=sqlite` and `MIMISBRUNNR_SQLITE_PATH=/tmp/custom.db`
- WHEN configuration is loaded
- THEN `Driver` is `sqlite` AND `SQLitePath` is `/tmp/custom.db`

---

### Requirement: REQ-205 Unimplemented Drivers Rejected at Open

Drivers not implemented in this version (including `postgres`) MUST be rejected in `storage.Open` with a clear error. The system MUST NOT silently fall back to SQLite when an explicit non-sqlite driver is requested.

#### Scenario: Postgres requested

- GIVEN `Config.Driver` is `postgres`
- WHEN `Open` is called
- THEN opening fails with an unsupported-driver error

---

### Requirement: REQ-207 Store Observation Persistence

`storage.Store` MUST expose methods to save, get by ID, list, and search observations, start/end sessions, and retrieve tiered context. MCP and HTTP layers MUST NOT access SQL directly; they use `Store` only.

#### Scenario: Save through store

- GIVEN a valid observation input without `topic_key`
- WHEN `SaveObservation` is called on the store
- THEN a persisted observation with assigned `ID` and timestamps is returned

#### Scenario: List through store

- GIVEN observations exist for project `p`
- WHEN `ListObservations` is called with `project=p`
- THEN matching observations are returned newest-first

#### Scenario: Search through store

- GIVEN observations exist for project `p` with searchable text
- WHEN `SearchObservations` is called with `project=p` and a matching `query`
- THEN relevant observations are returned ordered by FTS rank

#### Scenario: Start session through store

- GIVEN a valid project
- WHEN `StartSession` is called
- THEN an active session row is persisted

#### Scenario: Context through store

- GIVEN session-scoped observations exist
- WHEN `GetContext` is called with that `session_id`
- THEN tiered context entries are returned without full content

---

### Requirement: REQ-208 Observations Table Shape

The SQLite adapter MUST persist observations in table `observations` with columns: `id` (INTEGER PRIMARY KEY), `title`, `content`, `type`, `project`, `scope`, `topic_key` (nullable), `session_id` (nullable), `created_at`, `updated_at` (TEXT ISO-8601 UTC). A unique index MUST enforce `(project, scope, topic_key)` where `topic_key IS NOT NULL`.

#### Scenario: Unique topic key per project and scope

- GIVEN an observation exists with project `p`, scope `project`, topic_key `k`
- WHEN a second insert attempts the same triple without upsert semantics
- THEN the operation fails OR is handled by upsert per REQ-301

---

## Implementation References

- `internal/storage/store.go`
- `internal/storage/sqlite/sqlite.go`
- `internal/storage/sqlite/observations.go`
- `internal/storage/sqlite/search.go`
- `internal/storage/sqlite/fts.go`
- `internal/storage/sqlite/sessions.go`
- `internal/storage/sqlite/context.go`
- `internal/config/config.go`
