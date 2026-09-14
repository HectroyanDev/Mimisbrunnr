# Memory Observations (Delta)

## Purpose

Persist structured agent memory as observations that can be saved, retrieved, and listed by project without exposing MCP tools yet.

## ADDED Requirements

### Requirement: REQ-300 Observation Domain Model

The system MUST represent an observation with the following fields:

- `ID` (int64, assigned by storage on create)
- `Title` (non-empty string, max 200 characters)
- `Content` (non-empty string, max 32_000 characters)
- `Type` (non-empty string, max 64 characters; e.g. `decision`, `discovery`, `bugfix`, `manual`)
- `Project` (non-empty string, max 128 characters)
- `Scope` (`project` or `personal`; default `project` when omitted on create)
- `TopicKey` (optional string, max 128 characters; unique per project+scope when set)
- `CreatedAt` and `UpdatedAt` (UTC timestamps set by storage)

#### Scenario: Valid observation on create

- GIVEN required fields `title`, `content`, `type`, and `project` are provided
- WHEN the observation is validated for create
- THEN validation succeeds AND `scope` defaults to `project` if omitted

#### Scenario: Invalid empty title

- GIVEN `title` is empty or whitespace only
- WHEN the observation is validated for create
- THEN validation fails with a field error for `title`

#### Scenario: Invalid scope

- GIVEN `scope` is set to a value other than `project` or `personal`
- WHEN the observation is validated for create
- THEN validation fails with a field error for `scope`

---

### Requirement: REQ-301 Topic Key Upsert Semantics

When `topic_key` is provided on create, the system MUST upsert: if an active observation already exists for the same `project`, `scope`, and `topic_key`, the system MUST update that row's `title`, `content`, `type`, and `updated_at` instead of inserting a duplicate.

#### Scenario: First save with topic key

- GIVEN no observation exists for project `p`, scope `project`, topic_key `auth-model`
- WHEN an observation is saved with that topic key
- THEN a new row is created AND `created_at` and `updated_at` are set

#### Scenario: Second save with same topic key

- GIVEN an observation exists for project `p`, scope `project`, topic_key `auth-model`
- WHEN another observation is saved with the same project, scope, and topic_key
- THEN the existing row is updated AND `id` and `created_at` are unchanged AND `updated_at` advances

---

### Requirement: REQ-302 Read APIs Without Search

The system MUST support retrieving observations by ID and listing observations filtered by `project` and optional `scope`, ordered by `updated_at` descending. Full-text search and MCP exposure are out of scope for this capability.

#### Scenario: Get by ID

- GIVEN an observation with ID `42` exists
- WHEN it is retrieved by ID
- THEN the full observation is returned

#### Scenario: Get missing ID

- GIVEN no observation has ID `999`
- WHEN it is retrieved by ID
- THEN a not-found error is returned

#### Scenario: List by project

- GIVEN multiple observations exist for project `mimisbrunnr`
- WHEN observations are listed for that project with default limit
- THEN results are ordered by `updated_at` descending AND only observations for that project are returned
