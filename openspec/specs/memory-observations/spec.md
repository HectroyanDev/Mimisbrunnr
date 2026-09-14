# Memory Observations Specification

## Purpose

Persist structured agent memory as observations that can be saved, retrieved, listed, searched, and recalled in sessions by project. MCP exposure uses `mimir_save`, `mimir_get`, `mimir_list`, `mimir_search`, `mimir_session_start`, `mimir_session_end`, and `mimir_context`.

---

## Requirements

### Requirement: REQ-300 Observation Domain Model

The system MUST represent an observation with the following fields:

- `ID` (int64, assigned by storage on create)
- `Title` (non-empty string, max 200 characters)
- `Content` (non-empty string, max 32_000 characters)
- `Type` (non-empty string, max 64 characters; e.g. `decision`, `discovery`, `bugfix`, `manual`)
- `Project` (non-empty string, max 128 characters)
- `Scope` (`project` or `personal`; default `project` when omitted on create)
- `TopicKey` (optional string, max 128 characters; unique per project+scope when set)
- `SessionID` (optional string; links observation to an active session when set on save)
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

### Requirement: REQ-302 Read APIs

The system MUST support retrieving observations by ID and listing observations filtered by `project` and optional `scope`, ordered by `updated_at` descending. MCP tools `mimir_save`, `mimir_get`, and `mimir_list` expose these operations to agents.

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

---

### Requirement: REQ-303 Full-Text Search

The system MUST support searching observations by full-text query within a `project`, with optional `scope` filter and `limit` (default 20, max 100). Search MUST match tokenized terms in `title` and `content` and order results by relevance. MCP tool `mimir_search` exposes this operation.

#### Scenario: Search matches content

- GIVEN an observation exists for project `p` with content containing `JWT`
- WHEN observations are searched for project `p` with query `JWT`
- THEN that observation is returned

#### Scenario: Search respects scope

- GIVEN observations exist in project `p` for both scopes with matching content
- WHEN search is run with `scope=personal`
- THEN only personal-scope observations are returned

#### Scenario: Upsert updates search index

- GIVEN an observation is saved with `topic_key` and later upserted with new content
- WHEN search is run for terms only in the new content
- THEN the updated observation is returned AND old terms no longer match

---

### Requirement: REQ-400 Session Domain Model

The system MUST represent a session with:

- `ID` (string UUID, assigned on start)
- `Project` (non-empty, max 128 characters)
- `Scope` (`project` or `personal`; default `project`)
- `StartedAt` and `EndedAt` (UTC timestamps; `EndedAt` null while active)
- `Status` (`active` or `ended`)

#### Scenario: Start session

- GIVEN a valid `project`
- WHEN a session is started
- THEN a new session is created with `status=active` AND a unique `ID` is returned

#### Scenario: End session

- GIVEN an active session
- WHEN the session is ended with a non-empty summary
- THEN `status` becomes `ended` AND `ended_at` is set AND a `session_summary` observation is stored for that project

---

### Requirement: REQ-401 Session-Scoped Observations

When `session_id` is provided on save, the observation MUST be linked to that session. The session MUST exist and MUST be `active`.

#### Scenario: Save with active session

- GIVEN session `s` is active for project `p`
- WHEN an observation is saved with `session_id=s` and `project=p`
- THEN the observation is stored with `session_id=s`

#### Scenario: Save with ended session

- GIVEN session `s` has `status=ended`
- WHEN an observation is saved with `session_id=s`
- THEN validation fails with a field error for `session_id`

---

### Requirement: REQ-402 Tiered Context Retrieval

The system MUST support context retrieval that returns L0 (title) and L1 (snippet of content, max 500 characters) per observation, without full L2 content. Default limit MUST be 10 (max 50). When `session_id` is set, results MUST be limited to that session; otherwise recent observations for the `project` are returned ordered by `updated_at` descending.

#### Scenario: Context for session

- GIVEN observations linked to session `s`
- WHEN context is requested with `session_id=s`
- THEN only those observations are returned with `title` and `snippet` AND without full `content`

#### Scenario: Context without session

- GIVEN observations exist for project `p`
- WHEN context is requested with `project=p` and no `session_id`
- THEN recent project observations are returned as tiered entries

#### Scenario: Context includes readable text

- GIVEN at least one observation matches context
- WHEN context is retrieved
- THEN structured output includes a non-empty `context_text` suitable for agent prompts

---

### Requirement: REQ-403 Session Summary Type Convention

Ended sessions MUST persist their summary as an observation with `type=session_summary` and `topic_key` of the form `sessions/{session_id}/summary`.

#### Scenario: Summary topic key

- GIVEN session `abc` is ended with summary text
- WHEN the summary observation is stored
- THEN `topic_key` is `sessions/abc/summary` AND `type` is `session_summary`

---

## Implementation References

- `internal/memory/`
- `internal/memory/session.go`
- `internal/memory/context.go`
- `internal/storage/sqlite/observations.go`
- `internal/storage/sqlite/sessions.go`
- `internal/storage/sqlite/context.go`
