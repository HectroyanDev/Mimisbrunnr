package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/storage"
)

const (
	Name    = "mimisbrunnr"
	Version = "0.0.1"
)

// Tool names use the mimir_ prefix: the agent acts at Mímir's well (Mímisbrunnr).
const ToolMimirPing = "mimir_ping"

type server struct {
	store storage.Store
}

// NewServer builds the MCP server with stdio-ready tools.
func NewServer(store storage.Store) *sdk.Server {
	s := &server{store: store}
	mcpServer := sdk.NewServer(&sdk.Implementation{Name: Name, Version: Version}, nil)
	s.registerTools(mcpServer)
	return mcpServer
}

func (s *server) registerTools(mcpServer *sdk.Server) {
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirPing,
		Description: "Health check. Returns ok if the server process and storage are alive.",
	}, s.mimirPing)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirSave,
		Description: "Save an observation to agent memory. Upserts by topic_key when provided.",
	}, s.mimirSave)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirGet,
		Description: "Retrieve one observation by numeric ID.",
	}, s.mimirGet)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirList,
		Description: "List observations for a project, ordered by updated_at descending.",
	}, s.mimirList)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirSearch,
		Description: "Full-text search observations by title and content within a project.",
	}, s.mimirSearch)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirSessionStart,
		Description: "Begin an agent session at the well for a project.",
	}, s.mimirSessionStart)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirSessionEnd,
		Description: "End a session and persist a structured summary observation.",
	}, s.mimirSessionEnd)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        ToolMimirContext,
		Description: "Load recent tiered context (title and snippet) for a project or session.",
	}, s.mimirContext)
}

type mimirPingInput struct{}

type mimirPingOutput struct {
	Status string `json:"status" jsonschema:"ok if the server is alive"`
}

func (s *server) mimirPing(ctx context.Context, _ *sdk.CallToolRequest, _ mimirPingInput) (*sdk.CallToolResult, mimirPingOutput, error) {
	if err := s.store.Ping(ctx); err != nil {
		return nil, mimirPingOutput{}, fmt.Errorf("storage ping: %w", err)
	}
	return nil, mimirPingOutput{Status: "ok"}, nil
}
