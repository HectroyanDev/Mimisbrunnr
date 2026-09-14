# MCP Server Specification

## Purpose

Expose Mimisbrunnr as a local MCP server over stdio so agent clients (Cursor, Claude Desktop, VS Code) can spawn the binary as a subprocess and call tools. All agent-facing tools use the `mimir_` prefix: the agent acts at Mímir's well (Mímisbrunnr), distinct from generic `mem_*` naming used elsewhere.

---

## Requirements

### Requirement: REQ-100 MCP Server Identity and Stdio Transport

The system MUST run an MCP server named `mimisbrunnr` at version `0.0.1` using stdio transport (JSON-RPC over stdin/stdout).

#### Scenario: Client initializes the server

- GIVEN the `mimisbrunnr` binary is started with stdio transport
- WHEN an MCP client completes the initialize handshake
- THEN the server reports `serverInfo.name` as `mimisbrunnr` AND `serverInfo.version` as `0.0.1`

#### Scenario: Process speaks on stdio only

- GIVEN the binary is launched by an MCP client as a subprocess
- WHEN the client sends MCP messages on stdin
- THEN the server responds on stdout AND does not require a network listener for MCP in this version

---

### Requirement: REQ-101 Storage Injection

The MCP server MUST be constructed with a `storage.Store` dependency. The server MUST NOT open storage itself.

#### Scenario: Server receives store at construction

- GIVEN a valid `storage.Store` implementation
- WHEN `NewServer(store)` is called
- THEN the returned MCP server uses that store for tools that need persistence

---

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

### Requirement: REQ-102 Mimir Ping Tool Success

The system MUST register a tool named `mimir_ping` with no required arguments. When storage is healthy, calling `mimir_ping` MUST return structured output `{ "status": "ok" }`.

#### Scenario: Healthy storage

- GIVEN a `storage.Store` whose `Ping` succeeds
- WHEN a client calls the `mimir_ping` tool
- THEN the tool result is not an error AND structured content includes `status` equal to `"ok"`

---

### Requirement: REQ-103 Mimir Ping Tool Storage Failure

When `storage.Store.Ping` fails, the `mimir_ping` tool MUST surface an error to the client. It MUST NOT return success with `status: ok`.

#### Scenario: Storage ping fails

- GIVEN a `storage.Store` whose `Ping` returns an error
- WHEN a client calls the `mimir_ping` tool
- THEN the tool result is marked as an error

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

### Requirement: REQ-106 Mimir Get Tool

The system MUST register a tool named `mimir_get` with required argument `id` (positive int64). On success it MUST return the full observation. When the ID does not exist, the tool MUST surface an error.

#### Scenario: Get existing observation

- GIVEN an observation with ID `42` exists
- WHEN a client calls `mimir_get` with `id` 42
- THEN the tool result is not an error AND structured content matches the stored observation

#### Scenario: Get missing observation

- GIVEN no observation has ID `999`
- WHEN a client calls `mimir_get` with `id` 999
- THEN the tool result is marked as an error

---

### Requirement: REQ-107 Mimir List Tool

The system MUST register a tool named `mimir_list` with required argument `project`. Optional: `scope`, `limit` (default 20, max 100). Results MUST be ordered by `updated_at` descending.

#### Scenario: List by project

- GIVEN observations exist for project `mimisbrunnr`
- WHEN a client calls `mimir_list` with that project
- THEN the tool result is not an error AND structured content includes an `observations` array

---

### Requirement: REQ-108 Mimir Search Tool

The system MUST register a tool named `mimir_search` with required arguments `project` and `query`. Optional: `scope`, `limit` (default 20, max 100). Search MUST match against observation `title` and `content` using full-text search, ordered by relevance.

#### Scenario: Search finds matching observation

- GIVEN an observation exists with content containing `JWT`
- WHEN a client calls `mimir_search` with `project` and `query` `JWT`
- THEN the tool result is not an error AND structured content includes that observation in `observations`

#### Scenario: Search validation failure

- GIVEN `query` is empty or whitespace only
- WHEN a client calls `mimir_search`
- THEN the tool result is marked as an error

#### Scenario: All mimir tools registered

- GIVEN a freshly started MCP server with default registration
- WHEN a client lists tools
- THEN `mimir_ping`, `mimir_save`, `mimir_get`, `mimir_list`, `mimir_search`, `mimir_session_start`, `mimir_session_end`, and `mimir_context` are registered

---

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

---

## Implementation References

- `internal/mcp/server.go`
- `internal/mcp/tools_memory.go`
- `internal/mcp/tools_search.go`
- `internal/mcp/tools_session.go`
- `internal/mcp/tools_context.go`
- `cmd/mimisbrunnr/main.go`
