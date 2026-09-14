package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

const (
	ToolMimirSessionStart = "mimir_session_start"
	ToolMimirSessionEnd   = "mimir_session_end"
)

type mimirSessionStartInput struct {
	Project string `json:"project" jsonschema:"required,Project namespace for this session"`
	Scope   string `json:"scope,omitempty" jsonschema:"project or personal; defaults to project"`
}

type sessionOutput struct {
	SessionID string  `json:"session_id"`
	Project   string  `json:"project"`
	Scope     string  `json:"scope"`
	Status    string  `json:"status"`
	StartedAt string  `json:"started_at"`
	EndedAt   *string `json:"ended_at,omitempty"`
}

type mimirSessionEndInput struct {
	SessionID string `json:"session_id" jsonschema:"required,Active session to close"`
	Summary   string `json:"summary" jsonschema:"required,Structured session summary for long-term memory"`
}

func sessionFromMemory(sess memory.Session) sessionOutput {
	out := sessionOutput{
		SessionID: sess.ID,
		Project:   sess.Project,
		Scope:     sess.Scope,
		Status:    sess.Status,
		StartedAt: sess.StartedAt.UTC().Format(time.RFC3339Nano),
	}
	if sess.EndedAt != nil {
		ended := sess.EndedAt.UTC().Format(time.RFC3339Nano)
		out.EndedAt = &ended
	}
	return out
}

func (s *server) mimirSessionStart(ctx context.Context, _ *sdk.CallToolRequest, in mimirSessionStartInput) (*sdk.CallToolResult, sessionOutput, error) {
	if err := validateProject(in.Project); err != nil {
		return nil, sessionOutput{}, err
	}
	scope := in.Scope
	if scope != "" && scope != memory.ScopeProject && scope != memory.ScopePersonal {
		return nil, sessionOutput{}, &memory.ValidationError{Field: "scope", Message: "must be project or personal"}
	}

	sess, err := s.store.StartSession(ctx, in.Project, scope)
	if err != nil {
		return nil, sessionOutput{}, fmt.Errorf("start session: %w", err)
	}
	return nil, sessionFromMemory(sess), nil
}

func (s *server) mimirSessionEnd(ctx context.Context, _ *sdk.CallToolRequest, in mimirSessionEndInput) (*sdk.CallToolResult, sessionOutput, error) {
	sess, err := s.store.EndSession(ctx, memory.EndSessionInput{
		SessionID: in.SessionID,
		Summary:   in.Summary,
	})
	if err != nil {
		if errors.Is(err, memory.ErrSessionNotFound) || errors.Is(err, memory.ErrSessionEnded) {
			return nil, sessionOutput{}, err
		}
		return nil, sessionOutput{}, fmt.Errorf("end session: %w", err)
	}
	return nil, sessionFromMemory(sess), nil
}
