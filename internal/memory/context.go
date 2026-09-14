package memory

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ContextFilter selects tiered context for a project or session.
type ContextFilter struct {
	Project   string
	SessionID string
	Limit     int
}

// ContextEntry is L0/L1 context without full content.
type ContextEntry struct {
	ID        int64
	Title     string
	Snippet   string
	Type      string
	Project   string
	Scope     string
	TopicKey  *string
	SessionID *string
	UpdatedAt string
}

// ContextResult is tiered context for agents.
type ContextResult struct {
	ContextText  string
	Observations []ContextEntry
}

// ValidateContextFilter checks context parameters.
func ValidateContextFilter(f ContextFilter) error {
	project := strings.TrimSpace(f.Project)
	if project == "" {
		return &ValidationError{Field: "project", Message: "must not be empty"}
	}
	if utf8.RuneCountInString(project) > MaxProjectLen {
		return &ValidationError{Field: "project", Message: "exceeds maximum length"}
	}
	return nil
}

// NormalizeContextLimit applies default and max bounds.
func NormalizeContextLimit(limit int) int {
	if limit <= 0 {
		return DefaultContextLimit
	}
	if limit > MaxContextLimit {
		return MaxContextLimit
	}
	return limit
}

// TruncateSnippet returns L1 content capped at SnippetLen with ellipsis.
func TruncateSnippet(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	if utf8.RuneCountInString(content) <= SnippetLen {
		return content
	}
	runes := []rune(content)
	return string(runes[:SnippetLen]) + "..."
}

// FormatContextText builds readable context for agent prompts.
func FormatContextText(entries []ContextEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Mímir's well — recent context\n\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("## %s (%s)\n", e.Title, e.Type))
		if e.Snippet != "" {
			b.WriteString(e.Snippet)
			b.WriteByte('\n')
		}
		b.WriteString("\n---\n\n")
	}
	return strings.TrimSpace(b.String())
}
