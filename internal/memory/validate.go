package memory

import (
	"strings"
	"unicode/utf8"
)

// ValidateCreateInput checks create payload before persistence.
func ValidateCreateInput(in CreateInput) error {
	if err := requireNonEmpty("title", in.Title, MaxTitleLen); err != nil {
		return err
	}
	if err := requireNonEmpty("content", in.Content, MaxContentLen); err != nil {
		return err
	}
	if err := requireNonEmpty("type", in.Type, MaxTypeLen); err != nil {
		return err
	}
	if err := requireNonEmpty("project", in.Project, MaxProjectLen); err != nil {
		return err
	}

	scope := in.Scope
	if scope == "" {
		scope = ScopeProject
	}
	if scope != ScopeProject && scope != ScopePersonal {
		return &ValidationError{Field: "scope", Message: "must be project or personal"}
	}

	if in.TopicKey != "" {
		if utf8.RuneCountInString(in.TopicKey) > MaxTopicKeyLen {
			return &ValidationError{Field: "topic_key", Message: "exceeds maximum length"}
		}
	}

	if in.SessionID != "" {
		if utf8.RuneCountInString(strings.TrimSpace(in.SessionID)) > 64 {
			return &ValidationError{Field: "session_id", Message: "exceeds maximum length"}
		}
	}

	return nil
}

// NormalizeScope returns the effective scope for create input.
func NormalizeScope(scope string) string {
	if scope == "" {
		return ScopeProject
	}
	return scope
}

// NormalizeListLimit applies default and max bounds.
func NormalizeListLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}

func requireNonEmpty(field, value string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return &ValidationError{Field: field, Message: "must not be empty"}
	}
	if utf8.RuneCountInString(trimmed) > maxLen {
		return &ValidationError{Field: field, Message: "exceeds maximum length"}
	}
	return nil
}
