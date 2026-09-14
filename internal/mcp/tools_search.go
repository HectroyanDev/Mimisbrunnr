package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

const ToolMimirSearch = "mimir_search"

type mimirSearchInput struct {
	Project string `json:"project" jsonschema:"required,Project namespace to search within"`
	Query   string `json:"query" jsonschema:"required,Full-text search terms matched against title and content"`
	Scope   string `json:"scope,omitempty" jsonschema:"Filter by scope when set"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Max rows; default 20 max 100"`
}

func (s *server) mimirSearch(ctx context.Context, _ *sdk.CallToolRequest, in mimirSearchInput) (*sdk.CallToolResult, mimirListOutput, error) {
	filter := memory.SearchFilter{
		Project: in.Project,
		Query:   in.Query,
		Scope:   in.Scope,
		Limit:   in.Limit,
	}
	if err := memory.ValidateSearchFilter(filter); err != nil {
		return nil, mimirListOutput{}, err
	}

	rows, err := s.store.SearchObservations(ctx, filter)
	if err != nil {
		return nil, mimirListOutput{}, fmt.Errorf("search observations: %w", err)
	}

	out := mimirListOutput{Observations: make([]observationOutput, 0, len(rows))}
	for _, row := range rows {
		out.Observations = append(out.Observations, observationFromMemory(row))
	}
	return nil, out, nil
}
