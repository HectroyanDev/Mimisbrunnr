# Memory Observations (Delta)

## ADDED Requirements

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
