package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

const ToolMimirContext = "mimir_context"

type mimirContextInput struct {
	Project   string `json:"project" jsonschema:"required,Project namespace"`
	SessionID string `json:"session_id,omitempty" jsonschema:"Limit context to one session"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Max entries; default 10 max 50"`
}

type contextEntryOutput struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Snippet   string  `json:"snippet"`
	Type      string  `json:"type"`
	Project   string  `json:"project"`
	Scope     string  `json:"scope"`
	TopicKey  *string `json:"topic_key,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
	UpdatedAt string  `json:"updated_at"`
}

type mimirContextOutput struct {
	ContextText  string               `json:"context_text"`
	Observations []contextEntryOutput `json:"observations"`
}

func contextEntryFromMemory(e memory.ContextEntry) contextEntryOutput {
	return contextEntryOutput{
		ID:        e.ID,
		Title:     e.Title,
		Snippet:   e.Snippet,
		Type:      e.Type,
		Project:   e.Project,
		Scope:     e.Scope,
		TopicKey:  e.TopicKey,
		SessionID: e.SessionID,
		UpdatedAt: e.UpdatedAt,
	}
}

func (s *server) mimirContext(ctx context.Context, _ *sdk.CallToolRequest, in mimirContextInput) (*sdk.CallToolResult, mimirContextOutput, error) {
	filter := memory.ContextFilter{
		Project:   in.Project,
		SessionID: in.SessionID,
		Limit:     in.Limit,
	}
	if err := memory.ValidateContextFilter(filter); err != nil {
		return nil, mimirContextOutput{}, err
	}

	result, err := s.store.GetContext(ctx, filter)
	if err != nil {
		return nil, mimirContextOutput{}, fmt.Errorf("get context: %w", err)
	}

	out := mimirContextOutput{
		ContextText:  result.ContextText,
		Observations: make([]contextEntryOutput, 0, len(result.Observations)),
	}
	for _, row := range result.Observations {
		out.Observations = append(out.Observations, contextEntryFromMemory(row))
	}
	return nil, out, nil
}
