package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

const (
	ToolMimirSave = "mimir_save"
	ToolMimirGet  = "mimir_get"
	ToolMimirList = "mimir_list"
)

type mimirSaveInput struct {
	Title    string `json:"title" jsonschema:"required,Short title for the observation"`
	Content  string `json:"content" jsonschema:"required,Full observation body"`
	Type     string `json:"type" jsonschema:"required,Category such as decision discovery bugfix manual"`
	Project  string `json:"project" jsonschema:"required,Project namespace for this memory"`
	Scope    string `json:"scope,omitempty" jsonschema:"project or personal; defaults to project"`
	TopicKey  string `json:"topic_key,omitempty" jsonschema:"Optional stable key; upserts within project+scope"`
	SessionID string `json:"session_id,omitempty" jsonschema:"Optional active session to link this observation"`
}

type mimirGetInput struct {
	ID int64 `json:"id" jsonschema:"required,Observation ID to retrieve"`
}

type mimirListInput struct {
	Project string `json:"project" jsonschema:"required,Project namespace to list"`
	Scope   string `json:"scope,omitempty" jsonschema:"Filter by scope when set"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Max rows; default 20 max 100"`
}

type mimirListOutput struct {
	Observations []observationOutput `json:"observations"`
}

func (s *server) mimirSave(ctx context.Context, _ *sdk.CallToolRequest, in mimirSaveInput) (*sdk.CallToolResult, observationOutput, error) {
	payload := memory.CreateInput{
		Title:    in.Title,
		Content:  in.Content,
		Type:     in.Type,
		Project:  in.Project,
		Scope:    in.Scope,
		TopicKey:  in.TopicKey,
		SessionID: in.SessionID,
	}
	if err := memory.ValidateCreateInput(payload); err != nil {
		return nil, observationOutput{}, err
	}
	payload.Scope = memory.NormalizeScope(payload.Scope)

	obs, err := s.store.SaveObservation(ctx, payload)
	if err != nil {
		return nil, observationOutput{}, fmt.Errorf("save observation: %w", err)
	}
	return nil, observationFromMemory(obs), nil
}

func (s *server) mimirGet(ctx context.Context, _ *sdk.CallToolRequest, in mimirGetInput) (*sdk.CallToolResult, observationOutput, error) {
	if in.ID <= 0 {
		return nil, observationOutput{}, &memory.ValidationError{Field: "id", Message: "must be positive"}
	}

	obs, err := s.store.GetObservation(ctx, in.ID)
	if err != nil {
		if errors.Is(err, memory.ErrNotFound) {
			return nil, observationOutput{}, err
		}
		return nil, observationOutput{}, fmt.Errorf("get observation: %w", err)
	}
	return nil, observationFromMemory(obs), nil
}

func (s *server) mimirList(ctx context.Context, _ *sdk.CallToolRequest, in mimirListInput) (*sdk.CallToolResult, mimirListOutput, error) {
	if err := validateProject(in.Project); err != nil {
		return nil, mimirListOutput{}, err
	}

	rows, err := s.store.ListObservations(ctx, memory.ListFilter{
		Project: in.Project,
		Scope:   in.Scope,
		Limit:   memory.NormalizeListLimit(in.Limit),
	})
	if err != nil {
		return nil, mimirListOutput{}, fmt.Errorf("list observations: %w", err)
	}

	out := mimirListOutput{Observations: make([]observationOutput, 0, len(rows))}
	for _, row := range rows {
		out.Observations = append(out.Observations, observationFromMemory(row))
	}
	return nil, out, nil
}

func validateProject(project string) error {
	if strings.TrimSpace(project) == "" {
		return &memory.ValidationError{Field: "project", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(strings.TrimSpace(project)) > memory.MaxProjectLen {
		return &memory.ValidationError{Field: "project", Message: "exceeds maximum length"}
	}
	return nil
}
