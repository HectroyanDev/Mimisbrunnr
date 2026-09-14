package mcp

import (
	"time"

	"github.com/Gentleman-Programming/mimisbrunnr/internal/memory"
)

type observationOutput struct {
	ID        int64   `json:"id" jsonschema:"Observation ID assigned by storage"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      string  `json:"type"`
	Project   string  `json:"project"`
	Scope     string  `json:"scope"`
	TopicKey  *string `json:"topic_key,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
	CreatedAt string  `json:"created_at" jsonschema:"UTC timestamp (RFC3339Nano)"`
	UpdatedAt string  `json:"updated_at" jsonschema:"UTC timestamp (RFC3339Nano)"`
}

func observationFromMemory(obs memory.Observation) observationOutput {
	out := observationOutput{
		ID:        obs.ID,
		Title:     obs.Title,
		Content:   obs.Content,
		Type:      obs.Type,
		Project:   obs.Project,
		Scope:     obs.Scope,
		TopicKey:  obs.TopicKey,
		SessionID: obs.SessionID,
		CreatedAt: obs.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: obs.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	return out
}
