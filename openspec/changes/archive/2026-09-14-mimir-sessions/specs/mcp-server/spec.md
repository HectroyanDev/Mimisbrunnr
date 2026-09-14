# MCP Server (Delta)

## MODIFIED Requirements

### Requirement: REQ-105 Mimir Tool Naming Convention

All MCP tools exposed to agents MUST use the `mimir_` prefix (lowercase, underscore-separated action). Tools MUST NOT use the `mem_` prefix.

| Tool | Purpose |
|------|---------|
| `mimir_ping` | Health check (process + storage) |
| `mimir_save` | Persist an observation |
| `mimir_get` | Retrieve one observation by ID |
| `mimir_list` | List observations for a project |
| `mimir_search` | Full-text search across observations |
| `mimir_session_start` | Begin an agent session at the well |
| `mimir_session_end` | End session and store structured summary |
| `mimir_context` | Recent tiered context (title + snippet) |

#### Scenario: Tool names are prefixed

- GIVEN the MCP server registers a tool for agents
- WHEN a client lists tools
- THEN every tool name starts with `mimir_`

---

### Requirement: REQ-104 Mimir Save Tool

The system MUST register a tool named `mimir_save` with required arguments `title`, `content`, `type`, and `project`. Optional arguments: `scope` (defaults to `project`), `topic_key` (upsert within project+scope), `session_id` (link to active session). On success it MUST return the saved observation including assigned `id` and timestamps.

#### Scenario: Save new observation

- GIVEN valid create fields
- WHEN a client calls `mimir_save`
- THEN the tool result is not an error AND structured content includes the observation fields

#### Scenario: Save validation failure

- GIVEN `title` is empty or whitespace only
- WHEN a client calls `mimir_save`
- THEN the tool result is marked as an error

#### Scenario: Save linked to session

- GIVEN an active session `s`
- WHEN a client calls `mimir_save` with `session_id=s`
- THEN the saved observation is associated with session `s`

---

## ADDED Requirements

### Requirement: REQ-109 Mimir Session Start Tool

The system MUST register `mimir_session_start` with required `project` and optional `scope`. On success it MUST return `session_id`, `project`, `scope`, and `started_at`.

#### Scenario: Start session via MCP

- GIVEN a valid project
- WHEN a client calls `mimir_session_start`
- THEN the tool result is not an error AND includes `session_id`

---

### Requirement: REQ-110 Mimir Session End Tool

The system MUST register `mimir_session_end` with required `session_id` and `summary`. The session MUST transition to `ended` and a `session_summary` observation MUST be stored.

#### Scenario: End active session

- GIVEN an active session
- WHEN a client calls `mimir_session_end` with a non-empty summary
- THEN the tool result is not an error AND the session is no longer active

---

### Requirement: REQ-111 Mimir Context Tool

The system MUST register `mimir_context` with required `project`, optional `session_id`, and optional `limit` (default 10, max 50). Results MUST include `context_text` and tiered `observations` without full content.

#### Scenario: Context for project

- GIVEN observations exist for project `p`
- WHEN a client calls `mimir_context` with `project=p`
- THEN the tool returns tiered observations and `context_text`
