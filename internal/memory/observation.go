package memory

import "time"

const (
	ScopeProject  = "project"
	ScopePersonal = "personal"

	MaxTitleLen    = 200
	MaxContentLen  = 32_000
	MaxTypeLen     = 64
	MaxProjectLen  = 128
	MaxTopicKeyLen = 128

	DefaultListLimit = 20
	MaxListLimit     = 100
)

// Observation is a persisted agent memory entry.
type Observation struct {
	ID        int64
	Title     string
	Content   string
	Type      string
	Project   string
	Scope     string
	TopicKey  *string
	SessionID *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateInput is the payload for saving an observation.
type CreateInput struct {
	Title    string
	Content  string
	Type     string
	Project  string
	Scope    string
	TopicKey  string
	SessionID string
}

// ListFilter selects observations for a project.
type ListFilter struct {
	Project string
	Scope   string
	Limit   int
}
