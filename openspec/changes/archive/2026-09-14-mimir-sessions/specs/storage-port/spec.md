# Storage Port (Delta)

## MODIFIED Requirements

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

### Requirement: REQ-207 Store Observation Persistence

`storage.Store` MUST expose methods to save, get by ID, list, search observations, start/end sessions, and retrieve tiered context. MCP and HTTP layers MUST NOT access SQL directly; they use `Store` only.

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
