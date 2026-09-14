package memory

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	SessionStatusActive = "active"
	SessionStatusEnded  = "ended"

	TypeSessionSummary = "session_summary"

	DefaultContextLimit = 10
	MaxContextLimit     = 50
	SnippetLen          = 500
)

// Session is an agent work period at the well.
type Session struct {
	ID        string
	Project   string
	Scope     string
	Status    string
	StartedAt time.Time
	EndedAt   *time.Time
}

// EndSessionInput closes a session with a summary for long-term memory.
type EndSessionInput struct {
	SessionID string
	Summary   string
}

// ValidateEndSessionInput checks session end payload.
func ValidateEndSessionInput(in EndSessionInput) error {
	if strings.TrimSpace(in.SessionID) == "" {
		return &ValidationError{Field: "session_id", Message: "must not be empty"}
	}
	summary := strings.TrimSpace(in.Summary)
	if summary == "" {
		return &ValidationError{Field: "summary", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(summary) > MaxContentLen {
		return &ValidationError{Field: "summary", Message: "exceeds maximum length"}
	}
	return nil
}

// SessionSummaryTopicKey returns the canonical topic_key for a session summary.
func SessionSummaryTopicKey(sessionID string) string {
	return "sessions/" + strings.TrimSpace(sessionID) + "/summary"
}
