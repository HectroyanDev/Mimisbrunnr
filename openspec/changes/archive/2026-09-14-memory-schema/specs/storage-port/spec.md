# Storage Port (Delta)

## REMOVED Requirements

### Requirement: REQ-206 No Memory Schema in Baseline

**Reason**: Memory observations are now part of the product; bootstrap-only schema is superseded by migration v2.

**Migration**: Opening storage on an existing v1 database runs migration v2 and creates the `observations` table. `schema_meta.version` becomes `2`.

---

## MODIFIED Requirements

### Requirement: REQ-203 Schema Migration v1

On open, the SQLite adapter MUST ensure table `schema_meta(version INTEGER NOT NULL)` exists. The adapter MUST apply sequential migrations until the database reaches the current schema version. `Ping` MUST verify that `schema_meta.version` equals the current version (`2` after this change).

#### Scenario: Fresh database

- GIVEN a new SQLite file
- WHEN the adapter opens and `Ping` is called
- THEN `schema_meta` contains version `2` AND the `observations` table exists

#### Scenario: Upgrade from version 1

- GIVEN a database at `schema_meta.version = 1` with only bootstrap tables
- WHEN the adapter opens
- THEN migration v2 runs AND `schema_meta.version` becomes `2` AND existing v1 data is preserved

#### Scenario: Reopen existing database

- GIVEN a database already migrated to version `2`
- WHEN the adapter is closed and reopened on the same path
- THEN `Ping` still succeeds without reinitializing to a different version

#### Scenario: Unexpected schema version

- GIVEN `schema_meta` contains a version greater than the binary supports
- WHEN `Ping` is called
- THEN an error is returned indicating unsupported schema version

---

## ADDED Requirements

### Requirement: REQ-207 Store Observation Persistence

`storage.Store` MUST expose methods to save, get by ID, and list observations. MCP and HTTP layers MUST NOT access SQL directly; they use `Store` only.

#### Scenario: Save through store

- GIVEN a valid observation input without `topic_key`
- WHEN `SaveObservation` is called on the store
- THEN a persisted observation with assigned `ID` and timestamps is returned

#### Scenario: List through store

- GIVEN observations exist for project `p`
- WHEN `ListObservations` is called with `project=p`
- THEN matching observations are returned newest-first

---

### Requirement: REQ-208 Observations Table Shape

The SQLite adapter MUST persist observations in table `observations` with columns: `id` (INTEGER PRIMARY KEY), `title`, `content`, `type`, `project`, `scope`, `topic_key` (nullable), `created_at`, `updated_at` (TEXT ISO-8601 UTC). A unique index MUST enforce `(project, scope, topic_key)` where `topic_key IS NOT NULL`.

#### Scenario: Unique topic key per project and scope

- GIVEN an observation exists with project `p`, scope `project`, topic_key `k`
- WHEN a second insert attempts the same triple without upsert semantics
- THEN the operation fails OR is handled by upsert per REQ-301
